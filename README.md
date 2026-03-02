# 🚀 NEXUS - Plataforma Resiliente

Bem-vindo ao repositório oficial do projeto **NEXUS**. Este projeto é uma API de alto desempenho construída em linguagem **Go**, empacotada em múltiplos estágios no **Docker** e pensada para rodar em clusters **Kubernetes** com foco em FinOps, Alta Disponibilidade e Observabilidade profunda.

---

## 🏗️ Arquitetura e Tecnologias

- **Linguagem Principal**: Go (Golang) 1.24
- **Framework Web**: Gin
- **Containerização**: Docker (Multi-stage build gerando imagens minimalistas com *scratch*)
- **Orquestração**: Kubernetes (Deployments, Services tipo LoadBalancer)
- **Observabilidade (Cérebro)**: 
  - OpenTelemetry (Traces)
  - Stack LGTM (Loki, Grafana, Tempo, Prometheus/Mimir)

---

## 📂 Estrutura do Projeto

```text
NEXUS/
│
├── apps/
│   └── api-nexus/         # Código-fonte da API em Go
│       ├── main.go        # Lógica de negócio, rotas e instrumentação OTel
│       ├── Dockerfile     # Build otimizado para o cluster
│       └── go.mod         # Gerenciamento de dependências
│
├── infra/
│   └── k8s/               # Manifesto Declarativo (IaC) do Kubernetes
│       ├── deployment.yaml # Regras de réplicas e limites de hardware (FinOps)
│       └── service.yaml    # Exposição do LoadBalancer
│
├── observability/
│   └── values.yaml        # Configuração helm chart para stack LGTM
│
├── .gitignore             # Ignorar arquivos temporários/binários
├── SETUP.md               # Guia Rápido de Instalação e Recuperação
└── README.md              # Documentação principal
```

---

## ⚙️ Como Começar / Recuperar o Ambiente

Se você formatou sua máquina ou é um novo desenvolvedor na equipe, confira nosso [Guia de Setup e Instalação (SETUP.md)](./SETUP.md) para restaurar todo o ambiente local (Docker + K8s + Grafana) em menos de 5 minutos, do zero.

---

## 📊 A Magia da Observabilidade no Grafana

O projeto NEXUS implementa automaticamente a exportação de **Rastros (Traces)** distribuídos. Cada requisição enviada para a API gera um histórico exato de qual método foi executado e em qual réplica do Kubernetes o processamento aconteceu, utilizando o **Tempo** do Grafana.

Isso garante que gargalos de banco de dados ou funções lentas sejam visualizadas graficamente (Gráfico de Gantt) com exatidão no Painel do Engenheiro.

---

*Desenvolvido e documentado pelo Arquiteto Maicon.*
