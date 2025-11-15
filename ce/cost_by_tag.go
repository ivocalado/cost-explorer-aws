// mycostexplorer/cost_by_tag.go
package mycostexplorer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// GetLastNDaysTotalByTag retorna o custo total dos últimos N dias filtrado por tag
// Se days <= 0, retorna os custos do mês atual
// tagKey: nome da tag (ex: "Environment", "Project", "Team")
// tagValue: valor da tag para filtrar (ex: "Production", "MyApp", "DevOps")
func GetLastNDaysTotalByTag(ctx context.Context, ce *costexplorer.Client, days int, tagKey, tagValue string) error {
	var start, end string

	if days <= 0 {
		start, end = currentMonthRange()
		fmt.Printf("Consultando custo total do mês atual filtrado por tag %s=%s\n", tagKey, tagValue)
	} else {
		start, end = lastNDaysRange(days)
		fmt.Printf("Consultando custo total dos últimos %d dias filtrado por tag %s=%s\n", days, tagKey, tagValue)
	}
	fmt.Printf("Período: %s até %s\n\n", start, end)

	metric := "UnblendedCost"

	// Cria o filtro de tag
	filter := &types.Expression{
		Tags: &types.TagValues{
			Key:    aws.String(tagKey),
			Values: []string{tagValue},
		},
	}

	out, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityMonthly, // Total do período
		Metrics:     []string{metric},
		Filter:      filter,
	})
	if err != nil {
		return fmt.Errorf("GetCostAndUsage: %w", err)
	}

	if len(out.ResultsByTime) == 0 {
		fmt.Println("Nenhum dado de custo retornado.")
		return nil
	}

	// Calcula o total
	var totalCost float64
	for _, result := range out.ResultsByTime {
		mv, ok := result.Total[metric]
		if !ok || mv.Amount == nil {
			continue
		}

		var cost float64
		fmt.Sscanf(aws.ToString(mv.Amount), "%f", &cost)
		totalCost += cost
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("Custo Total com tag %s=%s: %.2f USD\n", tagKey, tagValue, totalCost)
	fmt.Printf("═══════════════════════════════════════════════════════════\n")

	return nil
}

// GetLastNDaysByTagGrouped retorna os custos dos últimos N dias filtrado por tag,
// com detalhamento diário e por serviço
func GetLastNDaysByTagGrouped(ctx context.Context, ce *costexplorer.Client, days int, tagKey, tagValue string) error {
	var start, end string

	if days <= 0 {
		start, end = currentMonthRange()
		fmt.Printf("Consultando custos diários por serviço do mês atual filtrado por tag %s=%s\n", tagKey, tagValue)
	} else {
		start, end = lastNDaysRange(days)
		fmt.Printf("Consultando custos diários por serviço dos últimos %d dias filtrado por tag %s=%s\n", days, tagKey, tagValue)
	}
	fmt.Printf("Período: %s até %s\n\n", start, end)

	metric := "UnblendedCost"

	// Cria o filtro de tag
	filter := &types.Expression{
		Tags: &types.TagValues{
			Key:    aws.String(tagKey),
			Values: []string{tagValue},
		},
	}

	out, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{metric},
		Filter:      filter,
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

	if len(out.ResultsByTime) == 0 {
		fmt.Println("Nenhum dado de custo retornado para esta tag.")
		return nil
	}

	var grandTotal float64

	// Processa cada dia
	for _, day := range out.ResultsByTime {
		startDate := aws.ToString(day.TimePeriod.Start)
		endDate := aws.ToString(day.TimePeriod.End)

		fmt.Printf("📅 Período: %s a %s\n", startDate, endDate)
		fmt.Println("───────────────────────────────────────────────────────────")

		if len(day.Groups) == 0 {
			fmt.Println("  Nenhum serviço com custo neste dia")
			fmt.Println()
			continue
		}

		var totalDay float64
		for _, group := range day.Groups {
			serviceName := "Unknown"
			if len(group.Keys) > 0 {
				serviceName = group.Keys[0]
			}

			mv, ok := group.Metrics[metric]
			if !ok || mv.Amount == nil {
				continue
			}

			amount := aws.ToString(mv.Amount)
			unit := aws.ToString(mv.Unit)

			var cost float64
			fmt.Sscanf(amount, "%f", &cost)
			totalDay += cost

			if cost > 0.01 {
				fmt.Printf("  %-40s: %10s %s\n", serviceName, amount, unit)
			}
		}

		grandTotal += totalDay
		fmt.Printf("  %s\n", "───────────────────────────────────────────────────────────")
		fmt.Printf("  %-40s: %10.2f USD\n", "TOTAL DO DIA", totalDay)
		fmt.Println()
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("TOTAL GERAL (tag %s=%s): %.2f USD\n", tagKey, tagValue, grandTotal)
	fmt.Printf("═══════════════════════════════════════════════════════════\n")

	return nil
}
