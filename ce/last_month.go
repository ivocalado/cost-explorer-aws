// mycostexplorer/last_month.go
package mycostexplorer

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func lastClosedMonthRange() (start, end string) {
	now := time.Now().UTC()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfLastMonth := firstOfThisMonth.AddDate(0, -1, 0)

	layout := "2006-01-02"
	return firstOfLastMonth.Format(layout), firstOfThisMonth.Format(layout)
}

func GetLastMonthTotalCost(ctx context.Context, ce *costexplorer.Client) error {
	start, end := lastClosedMonthRange()
	fmt.Printf("Consultando custos de %s até %s (end exclusivo)\n", start, end)

	metric := "UnblendedCost"

	out, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{metric},
	})
	if err != nil {
		return fmt.Errorf("GetCostAndUsage: %w", err)
	}

	if len(out.ResultsByTime) == 0 {
		fmt.Println("Nenhum dado de custo retornado.")
		return nil
	}

	day := out.ResultsByTime[0]

	mv, ok := day.Total[metric]
	if !ok {
		fmt.Printf("Nenhum valor para %s no resultado. Keys disponíveis: ", metric)
		for k := range day.Total {
			fmt.Printf("%s ", k)
		}
		fmt.Println()
		return nil
	}

	fmt.Printf("Custo total no período: %s %s\n",
		aws.ToString(mv.Amount),
		aws.ToString(mv.Unit),
	)
	return nil
}
