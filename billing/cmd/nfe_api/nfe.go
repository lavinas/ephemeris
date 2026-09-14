package main

const soapEnvelope = `
<?xml version="1.0" encoding="utf-8"?>
<soapenv:Envelope xmlns:soapenv="http://xmlsoap.org" xmlns:v1="http://www.prefeitura.sp.gov.br/nfe">
   <soapenv:Header/>
   <soapenv:Body>
      <v1:EnvioRPS>
         <v1:VersaoSchema>2</v1:VersaoSchema>
         <v1:MensagemXML><![CDATA[
            <PedidoEnvioRPS xmlns="http://www.prefeitura.sp.gov.br/nfe" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.prefeitura.sp.gov.br/nfe loteRPS_v2.xsd">
			<Cabecalho Versao="2">
                <CPFCNPJRemetente><CNPJ>12345678000190</CNPJ></CPFCNPJRemetente>
            </Cabecalho>
            <RPS>
                <Assinatura>ASSINATURA_DO_RPS_AQUI</Assinatura>
                <ChaveRPS>
                    <InscricaoPrestador>12345678</InscricaoPrestador>
                    <SerieRPS>AAAA</SerieRPS>
                    <NumeroRPS>101</NumeroRPS>
                </ChaveRPS>
                <TipoRPS>RPS</TipoRPS>
                <DataEmissao>2026-08-04</DataEmissao>
                <StatusRPS>N</StatusRPS>
                <TributacaoRPS>T</TributacaoRPS>
                <ValorServicos>150.00</ValorServicos>
                <CodigoServico>02801</CodigoServico>
                <AliquotaServico>0.05</AliquotaServico>
                <CpfCnpjTomador><CNPJ>98765432000111</CNPJ></CpfCnpjTomador>
                <Discriminacao>Servico de desenvolvimento de software prestado em Sao Paulo.</Discriminacao>
            </RPS>
               <!-- O bloco acima deve ser assinado digitalmente incluindo a tag Signature aqui -->
            </PedidoEnvioRPS>
         ]]></v1:MensagemXML>
      </v1:EnvioRPS>
   </soapenv:Body>
</soapenv:Envelope>`

const xmlRawTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<PedidoEnvioLoteRPS xmlns="http://prefeitura.sp.gov.br">
  <Cabecalho Versao="1">
    <CPFCNPJRemetente>
      <CNPJ>{{.CNPJRemetente}}</CNPJ>
    </CPFCNPJRemetente>
    <
    <QuantidadeRPS>{{.QuantidadeRPS}}</QuantidadeRPS>
    <ValorTotalServicos>{{.ValorTotalServicos}}</ValorTotalServicos>
  </Cabecalho>
  {{range .ListaRPS}}
  <RPS>
    <Assinatura>{{.Assinatura}}</Assinatura>
    <ChaveRPS>
      <InscricaoPrestador>{{.InscricaoPrestador}}</InscricaoPrestador>
      <SerieRPS>{{.Serie}}</SerieRPS>
      <NumeroRPS>{{.Numero}}</NumeroRPS>
    </ChaveRPS>
    <TipoRPS>RPS</TipoRPS>
    <DataEmissao>{{.DataEmissao}}</DataEmissao>
    <StatusRPS>N</StatusRPS>
    <TributacaoRPS>T</TributacaoRPS>
    <ValorServicos>{{.ValorServicos}}</ValorServicos>
    <ValorDeducoes>0.00</ValorDeducoes>
    <CodigoServico>{{.CodigoServico}}</CodigoServico>
    <AliquotaServico>{{.AliquotaServico}}</AliquotaServico>
    <ISSRetido>false</ISSRetido>
    <CpfCnpjTomador>
      <CNPJ>{{.CNPJTomador}}</CNPJ>
    </CpfCnpjTomador>
    <Discriminacao>{{.Discriminacao}}</Discriminacao>
  </RPS>
  {{end}}
</PedidoEnvioLoteRPS>`
