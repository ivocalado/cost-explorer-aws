// mycostexplorer/forecast.go
package mycostexplorer

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// GetCurrentMonthForecast retorna a previsão de custos para o mês atual
// e compara com o custo real do mês anterior
func GetCurrentMonthForecast(ctx context.Context, ce *costexplorer.Client) error {
	now := time.Now().UTC()
	layout := "2006-01-02"

	// Período do mês atual para forecast (do início ao fim do mês)
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := firstOfThisMonth.AddDate(0, 1, 0)
	endForecast := firstOfNextMonth.Format(layout)

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("📊 PREVISÃO DE CUSTOS - %s\n", firstOfThisMonth.Format("January 2006"))
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// 1. Busca os custos já acumulados no mês até hoje
	startMonthToDate := firstOfThisMonth.Format(layout)
	endMonthToDate := now.Format(layout)

	monthToDateOut, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startMonthToDate,
			End:   &endMonthToDate,
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
	})
	if err != nil {
		return fmt.Errorf("GetCostAndUsage (month to date): %w", err)
	}

	var monthToDateCost float64
	if len(monthToDateOut.ResultsByTime) > 0 {
		if mv, ok := monthToDateOut.ResultsByTime[0].Total["UnblendedCost"]; ok && mv.Amount != nil {
			fmt.Sscanf(aws.ToString(mv.Amount), "%f", &monthToDateCost)
		}
	}

	// 2. Busca a previsão total do mês (de hoje até o fim do mês)
	// GetCostForecast retorna a previsão TOTAL do período, incluindo o que já foi gasto
	// A API só aceita forecast começando de hoje ou futuro
	forecastStartPeriod := now.Format(layout)

	forecastOut, err := ce.GetCostForecast(ctx, &costexplorer.GetCostForecastInput{
		TimePeriod: &types.DateInterval{
			Start: &forecastStartPeriod,
			End:   &endForecast,
		},
		Metric:      types.MetricUnblendedCost,
		Granularity: types.GranularityMonthly,
	})
	if err != nil {
		return fmt.Errorf("GetCostForecast: %w", err)
	}

	var forecastedCost float64
	if forecastOut.Total != nil && forecastOut.Total.Amount != nil {
		fmt.Sscanf(aws.ToString(forecastOut.Total.Amount), "%f", &forecastedCost)
	}

	// Calcula quanto ainda será gasto: previsão total - já gasto
	forecastRemainingCost := forecastedCost - monthToDateCost

	// 3. Busca o custo real do mês anterior para comparação
	firstOfLastMonth := firstOfThisMonth.AddDate(0, -1, 0)
	startLastMonth := firstOfLastMonth.Format(layout)
	endLastMonth := firstOfThisMonth.Format(layout)

	lastMonthOut, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startLastMonth,
			End:   &endLastMonth,
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
	})
	if err != nil {
		return fmt.Errorf("GetCostAndUsage (last month): %w", err)
	}

	var lastMonthCost float64
	if len(lastMonthOut.ResultsByTime) > 0 {
		if mv, ok := lastMonthOut.ResultsByTime[0].Total["UnblendedCost"]; ok && mv.Amount != nil {
			fmt.Sscanf(aws.ToString(mv.Amount), "%f", &lastMonthCost)
		}
	}

	// 4. Calcula a diferença e percentual
	difference := forecastedCost - lastMonthCost
	var percentChange float64
	if lastMonthCost > 0 {
		percentChange = (difference / lastMonthCost) * 100
	}

	// 5. Exibe os resultados
	daysInMonth := firstOfNextMonth.Sub(firstOfThisMonth).Hours() / 24
	daysElapsed := now.Sub(firstOfThisMonth).Hours() / 24
	daysRemaining := firstOfNextMonth.Sub(now).Hours() / 24

	fmt.Printf("📅 Período do mês: %s até %s\n", startMonthToDate, endForecast)
	fmt.Printf("   Dias decorridos: %.0f de %.0f dias (%.0f dias restantes)\n\n", daysElapsed, daysInMonth, daysRemaining)

	fmt.Printf("Mês Anterior (%s):\n", firstOfLastMonth.Format("January 2006"))
	fmt.Printf("  Custo Real Total:        $ %12.2f USD\n\n", lastMonthCost)

	fmt.Printf("Mês Atual (%s):\n", firstOfThisMonth.Format("January 2006"))
	fmt.Printf("  Custo Acumulado (até hoje):  $ %12.2f USD\n", monthToDateCost)
	fmt.Printf("  Previsão (dias restantes):   $ %12.2f USD\n", forecastRemainingCost)
	fmt.Printf("  ─────────────────────────────────────────\n")
	fmt.Printf("  Previsão Total do Mês:       $ %12.2f USD\n\n", forecastedCost)

	fmt.Printf("───────────────────────────────────────────────────────────\n")
	fmt.Printf("Comparativo:\n")
	fmt.Printf("  Diferença:               $ %12.2f USD", difference)

	if difference >= 0 {
		fmt.Printf(" ⬆️  (+%.2f%%)\n", percentChange)
	} else {
		fmt.Printf(" ⬇️  (%.2f%%)\n", percentChange)
	}

	// Análise e recomendações
	fmt.Printf("\n")
	if percentChange > 10 {
		fmt.Printf("⚠️  ALERTA: Previsão indica aumento significativo de %.2f%% nos custos\n", percentChange)
	} else if percentChange > 5 {
		fmt.Printf("⚡ ATENÇÃO: Previsão indica aumento moderado de %.2f%% nos custos\n", percentChange)
	} else if percentChange < -5 {
		fmt.Printf("✅ Previsão indica redução de %.2f%% nos custos\n", -percentChange)
	} else {
		fmt.Printf("📊 Custos previstos estáveis (variação de %.2f%%)\n", percentChange)
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")

	return nil
}

// GetCurrentMonthForecastByService retorna a previsão de custos do mês atual agrupada por serviço
func GetCurrentMonthForecastByService(ctx context.Context, ce *costexplorer.Client) error {
	now := time.Now().UTC()
	layout := "2006-01-02"

	// Período do mês atual
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := firstOfThisMonth.AddDate(0, 1, 0)

	startMonth := firstOfThisMonth.Format(layout)
	endMonth := firstOfNextMonth.Format(layout)

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("📊 PREVISÃO DE CUSTOS POR SERVIÇO - %s\n", firstOfThisMonth.Format("January 2006"))
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")
	fmt.Printf("Período: %s até %s\n\n", startMonth, endMonth)

	// Busca custos atuais do mês com granularidade diária e agrupado por serviço
	actualOut, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startMonth,
			End:   &endMonth,
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("GetCostAndUsage: %w", err)
	}

	if len(actualOut.ResultsByTime) == 0 || len(actualOut.ResultsByTime[0].Groups) == 0 {
		fmt.Println("Nenhum dado de custo retornado.")
		return nil
	}

	// Processa e exibe custos por serviço
	var totalActual float64
	fmt.Println("Custos Acumulados até Agora por Serviço:")
	fmt.Println("───────────────────────────────────────────────────────────")

	for _, group := range actualOut.ResultsByTime[0].Groups {
		serviceName := "Unknown"
		if len(group.Keys) > 0 {
			serviceName = group.Keys[0]
		}

		mv, ok := group.Metrics["UnblendedCost"]
		if !ok || mv.Amount == nil {
			continue
		}

		var cost float64
		fmt.Sscanf(aws.ToString(mv.Amount), "%f", &cost)
		totalActual += cost

		if cost > 0.01 {
			fmt.Printf("  %-40s: $ %10.2f USD\n", serviceName, cost)
		}
	}

	fmt.Printf("───────────────────────────────────────────────────────────\n")
	fmt.Printf("  %-40s: $ %10.2f USD\n", "TOTAL ACUMULADO", totalActual)
	fmt.Printf("\n")

	// Busca a previsão total
	forecastOut, err := ce.GetCostForecast(ctx, &costexplorer.GetCostForecastInput{
		TimePeriod: &types.DateInterval{
			Start: &startMonth,
			End:   &endMonth,
		},
		Metric:      types.MetricUnblendedCost,
		Granularity: types.GranularityMonthly,
	})
	if err != nil {
		return fmt.Errorf("GetCostForecast: %w", err)
	}

	var forecastedTotal float64
	if forecastOut.Total != nil && forecastOut.Total.Amount != nil {
		fmt.Sscanf(aws.ToString(forecastOut.Total.Amount), "%f", &forecastedTotal)
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("PREVISÃO TOTAL PARA O MÊS:  $ %10.2f USD\n", forecastedTotal)

	remainingDays := firstOfNextMonth.Sub(now).Hours() / 24
	if remainingDays > 0 {
		estimatedRemaining := forecastedTotal - totalActual
		fmt.Printf("Estimativa Restante:        $ %10.2f USD (%.0f dias restantes)\n", estimatedRemaining, remainingDays)
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")

	return nil
}
