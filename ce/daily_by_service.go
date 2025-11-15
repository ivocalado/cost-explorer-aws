// mycostexplorer/daily_by_service.go
package mycostexplorer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// GetDailyByService retorna os custos diários do último mês, agrupados por serviço
func GetDailyByService(ctx context.Context, ce *costexplorer.Client) error {
	start, end := lastClosedMonthRange()
	fmt.Printf("Consultando custos diários por serviço de %s até %s\n\n", start, end)

	metric := "UnblendedCost"

	out, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{metric},
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
		fmt.Println("Nenhum dado de custo retornado.")
		return nil
	}

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

		// Ordena e exibe cada serviço
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

			// Converte para float para somar
			var cost float64
			fmt.Sscanf(amount, "%f", &cost)
			totalDay += cost

			if cost > 0.01 { // Mostra apenas custos significativos
				fmt.Printf("  %-40s: %10s %s\n", serviceName, amount, unit)
			}
		}

		fmt.Printf("  %s\n", "───────────────────────────────────────────────────────────")
		fmt.Printf("  %-40s: %10.2f USD\n", "TOTAL DO DIA", totalDay)
		fmt.Println()
	}

	return nil
}
