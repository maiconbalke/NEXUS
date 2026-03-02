# 🚀 Guia Oficial de Instalação Rápida - Projeto NEXUS

Este documento contém o passo a passo exato para reerguer toda a infraestrutura e código da API NEXUS do zero após formatar a máquina.

## 🛠️ 1. Pré-Requisitos

Antes de iniciar no computador recém formatado, instale as seguintes ferramentas:

1. **Docker Desktop** (Vá nas *Configurações* do Docker no Windows e marque a caixa para **Habilitar o Kubernetes**).
2. **Git** (Para baixar as pastas do NEXUS na nova máquina).
3. **Helm** (Gerenciador de pacotes da nuvem/Kubernetes). Via PowerShell como Administrador:
   ```powershell
   winget install Helm.Helm
   ```
   *(Feche o terminal e abra um novo após concluir a instalação para o Windows ler o comando).*

---

## 📦 2. Compilar a API (Fases 1 e 2)

Como nossa API é montada com o Go de forma super limpa, você não precisa instalar o Go no seu Windows novo (o Docker fará o trabalho interno usando o Go 1.24 de fundo).

1. Pelo terminal do VSCode, entre na pasta da API:
   ```bash
   cd apps/api-nexus
   ```
2. Mande o Docker construir a imagem e "salvar a base" chamada `nexus-api` no seu PC:
   ```bash
   docker build -t nexus-api:latest .
   ```

---

## ☸️ 3. Ligar os Motores (Fase 3 - Kubernetes)

Com a imagem salva no PC, agora vamos mandar o Kubernetes ligar 3 "trabalhadores independentes" (Réplicas) da API com as nossas travas de FinOps.

1. Volte para a raiz do projeto (pasta principal `NEXUS`):
   ```bash
   cd ../..
   ```
2. Dê a prancheta de configuração yaml para o Kubernetes criar as 3 cópias e o LoadBalancer 8081:
   ```bash
   kubectl apply -f infra/k8s/
   ```

---

## 🧠 4. Subir a Caixa Negra Analítica (Fases 4 e 5 - LGTM OTel)

Após o "Corpo" (a API) nascer, subiremos o "Cérebro" para inspecionar os logs e traces (Grafana, Tempo, Prometheus, Loki, Promtail) através das regras OTLP que criamos no `main.go`.

1. Adicione os catálogos oficiais do GitHub para o seu Helm novo:
   ```bash
   helm repo add grafana https://grafana.github.io/helm-charts
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm repo update
   ```
2. Instale o pacote base do Grafana + Prometheus + Loki pedindo para o helm ler nossas configurações de fábrica (`observability/values.yaml`):
   ```bash
   helm upgrade --install lgtm-stack grafana/loki-stack --namespace observability --create-namespace -f observability/values.yaml
   ```
3. Instale o receptor do **Tempo** (Nossa máquina de Raio-X OpenTelemetry Mágica):
   ```bash
   helm upgrade --install tempo grafana/tempo -n observability
   ```

---

## 🧪 5. A Hora do Jogo (Testando Tudo)

### 5.1 Fazer Pedidos na API
No PowerShell, mande requisições para criar tráfego no sistema:
```powershell
Invoke-RestMethod -Uri http://localhost:8081/order -Method Post -ContentType "application/json" -Body '{"id":"RESURREICAO_001", "amount":150000.99}'
```

### 5.2 Acessar Gráficos no Grafana
Faça o túnel invisível pelo terminal para trazer a porta secreta 80 pro seu Windows (na porta 3000):
```bash
kubectl port-forward svc/lgtm-stack-grafana 3000:80 -n observability
```
1. No navegador, acesse: **http://localhost:3000**
2. Login e Senha: `admin` / `admin`
3. Vá no botão da Bússola (**Explore**) > **Tempo** > Selecione o aplicativo em Busca > Clique no Span gerado para ver os metadados coloridos do seu Go na tela!

*Em algumas versões do Helm, confira sempre se o **Data Source** do Tempo na aba \`Connections\` do Grafana continua salvo com o URL \`http://tempo.observability.svc.cluster.local:3200\`.*
