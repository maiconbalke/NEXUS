// Package main é o ponto de entrada do programa Go.
package main

import (
	// "context" é fundamental para o OpenTelemetry: ele carrega o "rastro" (trace)
	// da requisição de uma função para outra sem que você precise passar tudo manualmente.
	// É a forma idiomática do Go de propagar valores como deadlines e dados de tracing.
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	// ─── OpenTelemetry: o "detector" de rastreamento ────────────────────────────
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// ─────────────────────────────────────────────────────────────────────────────
// Structs de Request / Response
// ─────────────────────────────────────────────────────────────────────────────

// OrderRequest define a estrutura esperada no corpo JSON da requisição POST /order.
type OrderRequest struct {
	Id     string  `json:"id"     binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// OrderResponse define o JSON de resposta enviado ao cliente.
type OrderResponse struct {
	OrderId       string `json:"orderId"`
	Status        string `json:"status"`
	ProcessedInMs int64  `json:"processedInMs"`
}

// ─────────────────────────────────────────────────────────────────────────────
// initTracer — Configura o OpenTelemetry
// ─────────────────────────────────────────────────────────────────────────────

// initTracer inicializa o TracerProvider global do OpenTelemetry.
// Retorna uma função "shutdown" que deve ser chamada ao encerrar o programa
// para garantir que todos os spans pendentes sejam exportados antes de sair.
func initTracer(ctx context.Context) (func(context.Context) error, error) {
	// 1. Criamos o EXPORTADOR OTLP gRPC que envia os dados para o Tempo.
	// O WithInsecure() diz que não vamos usar TLS criptografado (já que é na mesma rede local Kubernetes)
	// O WithEndpoint aponta para o serviço do Tempo na porta 4317 do gRPC OTLP.
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("tempo.observability.svc.cluster.local:4317"),
	)
	if err != nil {
		return nil, err
	}

	// 2. Criamos o RESOURCE: metadados que identificam ESTE serviço em todos os traces.
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("api-nexus"),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// 3. Criamos o TRACERPROVIDER com o exportador OTLP.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// 4. Registramos como global.
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// processOrder — Lógica de negócio com um Span de rastreamento
// ─────────────────────────────────────────────────────────────────────────────

// processOrder agora recebe um "ctx context.Context" como primeiro argumento.
// Esta é a convenção mais importante do Go: context SEMPRE é o primeiro parâmetro.
// O contexto carrega o "trace pai" da requisição HTTP para que o span desta função
// apareça DENTRO do trace da requisição no dashboard — formando uma árvore de spans.
func processOrder(ctx context.Context, order OrderRequest) OrderResponse {
	// otel.Tracer("api-nexus") obtém um Tracer com o nome do nosso componente.
	// O nome aparece nos dados do trace e ajuda a identificar qual parte do código gerou o span.
	tracer := otel.Tracer("api-nexus")

	// tracer.Start cria um novo SPAN — a unidade fundamental do tracing.
	// Um span representa uma operação com início e fim, como uma câmera gravando um clipe.
	//   - "ctx" é o contexto pai (trace da requisição HTTP).
	//   - "processOrder" é o nome do span que aparecerá no dashboard.
	// A função retorna um novo ctx (com este span como pai de futuros filhos) e o span em si.
	ctx, span := tracer.Start(ctx, "processOrder")
	// "defer span.End()" garante que o span é finalizado (e seu tempo registrado)
	// INDEPENDENTEMENTE de como a função retorna — com sucesso ou com erro.
	defer span.End()

	// Adicionamos ATRIBUTOS ao span: metadados de negócio que facilitam a depuração.
	// No dashboard você pode filtrar todos os traces de um pedido específico pelo order.id.
	span.SetAttributes(
		attribute.String("order.id", order.Id),
		attribute.Float64("order.amount", order.Amount),
	)

	// Simula o processamento com delay aleatório (100ms a 500ms).
	start := time.Now()
	delayMs := rand.Intn(401) + 100
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	elapsed := time.Since(start).Milliseconds()

	// Adicionamos o tempo de processamento ao span para correlacionar com os dashboards.
	span.SetAttributes(attribute.Int64("order.processing_ms", elapsed))

	return OrderResponse{
		OrderId:       order.Id,
		Status:        "processed",
		ProcessedInMs: elapsed,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// main — Ponto de entrada
// ─────────────────────────────────────────────────────────────────────────────

func main() {
	ctx := context.Background()
	// Inicializa o TracerProvider (agora apontando pro Tempo!)
	shutdown, err := initTracer(ctx)
	if err != nil {
		log.Fatalf("falha ao inicializar o tracer: %v", err)
	}
	defer shutdown(ctx)

	router := gin.Default()
	
	// Injeta o Middleware do OpenTelemetry para rastrear automaticamente todas as rotas do Gin.
	// Cada requisição HTTP de entrada vai gerar automaticamente um Span Raiz!
	router.Use(otelgin.Middleware("api-nexus"))

	router.POST("/order", func(c *gin.Context) {
		var req OrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "dados inválidos na requisição",
				"details": err.Error(),
			})
			return
		}

		// c.Request.Context() devolve o contexto da requisição HTTP atual.
		// Ele já contém informações de cancelamento e timeout do servidor.
		// Ao passá-lo para processOrder, "conectamos" o span filho ao trace raiz da requisição.
		response := processOrder(c.Request.Context(), req)

		c.JSON(http.StatusOK, response)
	})

	log.Println("✅ api-nexus rodando em :8080")
	router.Run(":8080")
}
