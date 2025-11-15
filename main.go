// main.go
package main

import (
	"context"
	mycostexplorer "example/ce"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

func newCostExplorerClient(ctx context.Context, region string, profile string) *costexplorer.Client {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	return costexplorer.NewFromConfig(cfg)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  cost-explorer <command> [options]

Commands:
  last-month-total          Mostra o custo total do último mês fechado
  daily-by-service          Mostra os custos diários agrupados por serviço do último mês
  last-n-daily-by-service   Mostra os custos diários agrupados por serviço dos últimos N dias
                            (ou do mês atual se N não for informado)
  cost-by-tag               Mostra o custo total dos últimos N dias filtrado por tag
  cost-by-tag-detailed      Mostra os custos diários por serviço filtrado por tag
  forecast                  Mostra a previsão de custos do mês atual com comparativo ao mês anterior
  forecast-by-service       Mostra a previsão de custos do mês atual agrupada por serviço

Use "cost-explorer <command> -h" para ver as opções de cada comando.
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {

	case "last-month-total":
		// flags específicas desse comando
		fs := flag.NewFlagSet("last-month-total", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		_ = fs.Parse(os.Args[2:])

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetLastMonthTotalCost(ctx, ce); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "daily-by-service":
		// flags específicas desse comando
		fs := flag.NewFlagSet("daily-by-service", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		_ = fs.Parse(os.Args[2:])

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetDailyByService(ctx, ce); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "last-n-daily-by-service":
		// flags específicas desse comando
		fs := flag.NewFlagSet("last-n-daily-by-service", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		days := fs.Int("days", 0, "Número de dias para consultar (0 ou omitido = mês atual)")
		_ = fs.Parse(os.Args[2:])

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetLastNDailyByService(ctx, ce, *days); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "cost-by-tag":
		// flags específicas desse comando
		fs := flag.NewFlagSet("cost-by-tag", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		days := fs.Int("days", 0, "Número de dias para consultar (0 ou omitido = mês atual)")
		tagKey := fs.String("tag-key", "", "Nome da tag para filtrar (ex: Environment, Project, Team)")
		tagValue := fs.String("tag-value", "", "Valor da tag para filtrar (ex: Production, MyApp)")
		_ = fs.Parse(os.Args[2:])

		if *tagKey == "" || *tagValue == "" {
			fmt.Fprintf(os.Stderr, "Erro: -tag-key e -tag-value são obrigatórios\n\n")
			fs.Usage()
			os.Exit(1)
		}

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetLastNDaysTotalByTag(ctx, ce, *days, *tagKey, *tagValue); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "cost-by-tag-detailed":
		// flags específicas desse comando
		fs := flag.NewFlagSet("cost-by-tag-detailed", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		days := fs.Int("days", 0, "Número de dias para consultar (0 ou omitido = mês atual)")
		tagKey := fs.String("tag-key", "", "Nome da tag para filtrar (ex: Environment, Project, Team)")
		tagValue := fs.String("tag-value", "", "Valor da tag para filtrar (ex: Production, MyApp)")
		_ = fs.Parse(os.Args[2:])

		if *tagKey == "" || *tagValue == "" {
			fmt.Fprintf(os.Stderr, "Erro: -tag-key e -tag-value são obrigatórios\n\n")
			fs.Usage()
			os.Exit(1)
		}

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetLastNDaysByTagGrouped(ctx, ce, *days, *tagKey, *tagValue); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "forecast":
		// flags específicas desse comando
		fs := flag.NewFlagSet("forecast", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		_ = fs.Parse(os.Args[2:])

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetCurrentMonthForecast(ctx, ce); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "forecast-by-service":
		// flags específicas desse comando
		fs := flag.NewFlagSet("forecast-by-service", flag.ExitOnError)
		profile := fs.String("profile", "", "AWS profile a ser usado (opcional, usa default se vazio)")
		region := fs.String("region", "us-east-1", "Região para o Cost Explorer (normalmente us-east-1)")
		_ = fs.Parse(os.Args[2:])

		ctx := context.Background()
		ce := newCostExplorerClient(ctx, *region, *profile)

		if err := mycostexplorer.GetCurrentMonthForecastByService(ctx, ce); err != nil {
			log.Fatalf("erro: %v", err)
		}

	case "-h", "--help", "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "Comando desconhecido: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}
