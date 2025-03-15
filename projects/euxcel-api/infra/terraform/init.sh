#!/bin/sh
set -euo pipefail

echo "$(date +'%Y-%m-%d %H:%M:%S') - Inicializando Terraform..."

if [ ! -f *.tf ]; then
  echo "No se encontraron archivos .tf en el directorio."
  exit 1
fi

# Inicializar Terraform
terraform init
if [ $? -ne 0 ]; then
  echo "Error durante terraform init. Verifica la configuración."
  exit 1
fi

echo "$(date +'%Y-%m-%d %H:%M:%S') - Validando configuración de Terraform..."
terraform validate
if [ $? -ne 0 ]; then
  echo "Error en la validación de la configuración de Terraform."
  exit 1
fi

echo "$(date +'%Y-%m-%d %H:%M:%S') - Planeando la aplicación de cambios..."
terraform plan -out=tfplan
if [ $? -ne 0 ]; then
  echo "Error durante terraform plan."
  exit 1
fi

echo "$(date +'%Y-%m-%d %H:%M:%S') - Aplicando configuración a Vault y Consul..."
terraform apply -auto-approve tfplan
if [ $? -ne 0 ]; then
  echo "Error durante terraform apply."
  exit 1
fi

rm -f tfplan
echo "$(date +'%Y-%m-%d %H:%M:%S') - Configuración aplicada correctamente. Vault y Consul están listos."
