package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	logoPath    = "./images/logo_amelia.png"
	htmlContent = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<style>
		/* Configuração base para o comportamento de página A4 */
		html, body { height: 100%; margin: 0; padding: 0; }
		body { font-family: Arial, sans-serif; color: #1a1a1a; line-height: 1.5; padding: 20px; box-sizing: border-box; position: relative; }
		
		.header-table { width: 100%; border-collapse: collapse; margin-bottom: 25px; }
		.company-logo { max-height: 85px; display: block; margin: 0 auto 15px auto; }
		.company-info { text-align: center; font-size: 13px; color: #333; }
		.company-name { font-size: 18px; font-weight: bold; color: #1a1a1a; margin-bottom: 0px; }
		
		/* Título de Recibo centralizado e limpo antes da primeira linha */
		.document-title { text-align: center; font-size: 24px; font-weight: bold; color: #1a1a1a; text-transform: uppercase; letter-spacing: 1px; margin-top: 5px; margin-bottom: 2px; }
		
		/* Subtítulo aproximado de RECIBO */
		.document-subtitle { text-align: center; font-size: 15px; font-weight: normal; color: #4a5568; margin-top: -5px; margin-bottom: 12px; }
		
		/* Linhas finas divisórias sem margens extras para controle total via padding */
		.divider { border-top: 1px solid #cbd5e1; margin-bottom: 0; }
		.divider-items { border-top: 1px solid #cbd5e1; margin-top: 0; margin-bottom: 25px; }
		
		/* Espaçamento simétrico acima e abaixo do bloco de dados */
		.details-table { width: 100%; border-collapse: collapse; font-size: 14px; margin-top: 0; }
		.details-td { width: 50%; vertical-align: top; padding-top: 12px; padding-bottom: 12px; }
		
		.label { font-weight: bold; color: #1a1a1a; }
		.val { color: #4a5568; }
		
		.text-black-bold { color: #1a1a1a !important; font-weight: bold !important; }
		
		/* Tabela configurada para expandir dinamicamente horizontalmente */
		.items-table { border-collapse: collapse; margin-bottom: 20px; margin-top: -15px; min-width: 100%; width: auto; }
		.items-table th { border-bottom: 2px solid #1a1a1a; padding: 10px 5px; font-size: 14px; font-weight: bold; text-align: left; }
		.items-table td { padding: 12px 5px; border-bottom: 1px solid #e2e8f0; font-size: 14px; color: #2d3748; }
		
		/* Impede a quebra de linha do texto da descrição */
		.desc-col { white-space: nowrap; }
		
		.total-row-td { border-bottom: none !important; padding-top: 8px !important; }
		
		/* Bloco de Informações Adicionais em linhas inteiras verticais */
		.extra-info-container { margin-top: 60px; width: 100%; }
		
		/* AJUSTE: Estilização do título em linha separada */
		.observacao-titulo { font-size: 14px; font-weight: bold; color: #1a1a1a; margin-bottom: 8px; width: 100%; }
		
		/* AJUSTE: Lista em formato bullet point e itálico suave ocupando a linha inteira */
		.observacao-lista { font-size: 13px; color: #4a5568; font-style: italic; width: 100%; margin-top: 0; padding-left: 20px; margin-bottom: 35px; }
		.observacao-lista li { margin-bottom: 6px; }
		
		/* Posicionado de forma limpa abaixo da lista */
		.signature-box { text-align: left; width: 100%; }
		.signature-title { font-size: 14px; font-weight: bold; color: #1a1a1a; font-style: normal; }
		
		/* Posiciona a frase e a última linha fixa no rodapé absoluto da página */
		.validation-footer { 
			position: absolute; 
			bottom: 20px; 
			left: 20px; 
			right: 20px; 
			text-align: center; 
			font-size: 11px; 
			color: #718096; 
			border-top: 1px solid #cbd5e1; 
			padding-top: 15px; 
		}
	</style>
</head>
<body>
	<!-- Topo com Logo e Dados da Empresa -->
	<table class="header-table">
		<tr>
			<td>
				<img src="{{.CidQRCode}}" class="company-logo" alt="Logo">
				<div class="company-info">
					<div class="company-name">Cardoso e Barbosa Serviços Musicais e Tecnologia LTDA</div>
					CNPJ: 27.928.875/0001-04<br>
					Email: financeiro@ameilacardoso.com.br<br>
					WhatsApp: (11) 98088-8399
				</div>
			</td>
		</tr>
	</table>

	<!-- Palavra RECIBO isolada e elegante no topo -->
	<div class="document-title">RECIBO</div>
	
	<!-- Subtítulo aproximado de RECIBO -->
	<div class="document-subtitle">Invoice #583</div>

	<!-- PRIMEIRA LINHA FINA -->
	<div class="divider"></div>

	<!-- Dados do Cliente e Identificação -->
	<table class="details-table">
		<tr>
			<td class="details-td">
				<span class="label">Ana Flavia de Andrade da Silva</span><br>
				<span class="label">CPF:</span> <span class="val">381.793.318-57</span><br>
				<span class="label">Email:</span> <span class="val">andrade.af22@gmail.com</span>
			</td>
			<td class="details-td" style="text-align: right;">
				<span class="label">Emissão :</span> <span class="val">06/07/2026</span><br>
				<span class="label">Vencimento:</span> <span class="val">10/07/2026</span><br>
				<span class="label">Pagamento:</span> <span class="val">10/07/2026</span>
			</td>
		</tr>
	</table>

	<!-- SEGUNDA LINHA FINA -->
	<div class="divider-items"></div>

	<!-- Descrição dos Valores -->
	<table class="items-table">
		<thead>
			<tr>
				<th>Descrição</th>
				<th style="text-align: center; width: 50px;">Quantidade</th>
				<th style="text-align: right; width: 80px;">Valor</th>
				<th style="text-align: right; width: 110px;">Total</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td class="desc-col">aulas de canto de 60 minutos em junho de 2026</td>
				<td style="text-align: center;">2</td>
				<td style="text-align: right;">R$ 110,00</td>
				<td style="text-align: right;">R$ 220,00</td>
			</tr>
			<tr>
				<td colspan="2" class="total-row-td"></td>
				<td style="text-align: right; font-weight: bold;" class="total-row-td">TOTAL</td>
				<td style="text-align: right; font-size: 15px;" class="text-black-bold total-row-td">R$ 220,00</td>
			</tr>
		</tbody>
	</table>

	<!-- BLOCO DE TEXTOS (Formatado com Título e marcador Bullet) -->
	<div class="extra-info-container">
		<div class="observacao-titulo">Notas:</div>
		<ul class="observacao-lista">
			<li>A nota fiscal de prestação de serviços correspondente a este pagamento será emitida e enviada eletronicamente de forma posterior.</li>
		</ul>
		
		<div class="signature-box">
			<div class="signature-title">Estúdio Amélia Cardoso</div>
		</div>
	</div>

	<!-- ÚLTIMA LINHA FINA E FRASE FIXADAS COMO RODAPÉ DA PÁGINA -->
	<div class="validation-footer">
		Este documento é uma representação do boleto pago e não possui validade fiscal.<br>
		Gerado por: Cardoso e Barbosa Serviços Musicais e Tecnologia LTDA
	</div>
</body>
</html>


`
)

func main() {
	fmt.Println("Lendo arquivo logo...")
	logoData, err := os.ReadFile(logoPath)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo de logo: %v", err)
	}

	// Converte a imagem para base64
	logoBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(logoData)

	// Substitui o placeholder no HTML pelo conteúdo base64 da imagem
	htmlContent := strings.ReplaceAll(htmlContent, "{{.CidQRCode}}", logoBase64)

	fmt.Println("Iniciando a geração do PDF...")

	// 1. CONFIGURAÇÃO CRUCIAL: Flags para ignorar GPU e Sandbox (Evita travamentos)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Headless,
	)

	// Cria o alocador com as opções definidas
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	// 2. TIMEOUT DE SEGURANÇA: Fecha após 15s se houver erro
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()
	ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)
	defer cancelTimeout()

	// htmlContent := `<h1>Teste</h1><p>PDF Gerado!</p>`

	var pdfBuffer []byte

	// 3. Execução: Injeção direta de conteúdo (Sem Navigate com strings perigosas)
	err = chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),

		// SUBSTITUI O SLEEP: Espera nativamente até que a tag body esteja 100% pronta
		chromedp.WaitReady("body", chromedp.ByQuery),

		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuffer, _, err = page.PrintToPDF().
				WithPaperWidth(8.27).   // Largura exata do A4
				WithPaperHeight(11.69). // Altura exata do A4
				WithMarginTop(0.59).    // Margem Superior (~1.5 cm)
				WithMarginBottom(0.59). // Margem Inferior (~1.5 cm)
				WithMarginLeft(0.59).   // Margem Esquerda (~1.5 cm)
				WithMarginRight(0.59).  // Margem Direita (~1.5 cm)
				WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	os.WriteFile("saida.pdf", pdfBuffer, 0644)
	fmt.Println("Sucesso! 'saida.pdf' gerado.")
}
