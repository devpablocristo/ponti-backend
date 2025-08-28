#!/bin/bash

# Script simple de pruebas en curl para el endpoint del Dashboard
# Ejecutar cuando el servidor esté corriendo en localhost:8080

BASE_URL="http://localhost:8080/api/v1/dashboard"

echo "🎯 Dashboard API - Pruebas Simples con curl"
echo "==========================================="
echo "Base URL: $BASE_URL"
echo ""

# Función para probar endpoint básico
test_basic() {
    echo "🧪 Probando endpoint básico..."
    echo "GET $BASE_URL"
    echo ""
    
    curl -s "$BASE_URL" | jq '.' 2>/dev/null || curl -s "$BASE_URL"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para probar filtro por customer_ids
test_customer() {
    echo "🧪 Probando filtro por customer_ids..."
    echo "GET $BASE_URL?customer_ids=1,2,3"
    echo ""
    
    curl -s "$BASE_URL?customer_ids=1,2,3" | jq '.' 2>/dev/null || curl -s "$BASE_URL?customer_ids=1,2,3"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para probar filtro por project_ids
test_project() {
    echo "🧪 Probando filtro por project_ids..."
    echo "GET $BASE_URL?project_ids=10,20"
    echo ""
    
    curl -s "$BASE_URL?project_ids=10,20" | jq '.' 2>/dev/null || curl -s "$BASE_URL?project_ids=10,20"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para probar filtro por campaign_ids
test_campaign() {
    echo "🧪 Probando filtro por campaign_ids..."
    echo "GET $BASE_URL?campaign_ids=100,200"
    echo ""
    
    curl -s "$BASE_URL?campaign_ids=100,200" | jq '.' 2>/dev/null || curl -s "$BASE_URL?campaign_ids=100,200"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para probar filtro por field_ids
test_field() {
    echo "🧪 Probando filtro por field_ids..."
    echo "GET $BASE_URL?field_ids=1000,2000"
    echo ""
    
    curl -s "$BASE_URL?field_ids=1000,2000" | jq '.' 2>/dev/null || curl -s "$BASE_URL?field_ids=1000,2000"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para probar combinación de filtros
test_combined() {
    echo "🧪 Probando combinación de filtros..."
    echo "GET $BASE_URL?customer_ids=1,2&project_ids=10&campaign_ids=100"
    echo ""
    
    curl -s "$BASE_URL?customer_ids=1,2&project_ids=10&campaign_ids=100" | jq '.' 2>/dev/null || curl -s "$BASE_URL?customer_ids=1,2&project_ids=10&campaign_ids=100"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para probar caso de error
test_error() {
    echo "🧪 Probando caso de error..."
    echo "GET $BASE_URL?customer_ids=abc"
    echo ""
    
    curl -s "$BASE_URL?customer_ids=abc" | jq '.' 2>/dev/null || curl -s "$BASE_URL?customer_ids=abc"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Función para mostrar menú
show_menu() {
    echo "📋 Menú de Pruebas"
    echo "=================="
    echo ""
    echo "1) Probar endpoint básico (sin filtros)"
    echo "2) Probar filtro por customer_ids"
    echo "3) Probar filtro por project_ids"
    echo "4) Probar filtro por campaign_ids"
    echo "5) Probar filtro por field_ids"
    echo "6) Probar combinación de filtros"
    echo "7) Probar caso de error"
    echo "8) Ejecutar todas las pruebas"
    echo "9) Salir"
    echo ""
    echo "💡 Para ejecutar el servidor: make run"
    echo ""
}

# Función para ejecutar todas las pruebas
run_all_tests() {
    echo "🚀 Ejecutando todas las pruebas..."
    echo ""
    
    test_basic
    test_customer
    test_project
    test_campaign
    test_field
    test_combined
    test_error
    
    echo "✅ Todas las pruebas completadas"
    echo ""
}

# Función principal
main() {
    while true; do
        show_menu
        read -p "Selecciona una opción (1-9): " choice
        
        case $choice in
            1) test_basic ;;
            2) test_customer ;;
            3) test_project ;;
            4) test_campaign ;;
            5) test_field ;;
            6) test_combined ;;
            7) test_error ;;
            8) run_all_tests ;;
            9) echo "👋 ¡Hasta luego!"; exit 0 ;;
            *) echo "❌ Opción inválida. Intenta de nuevo." ;;
        esac
        
        echo ""
        read -p "Presiona Enter para continuar..."
        echo ""
    done
}

# Ejecutar función principal
main
