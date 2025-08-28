#!/bin/bash

# Script de pruebas en curl para el endpoint del Dashboard
# Ejecutar cuando el servidor esté corriendo en localhost:8080

BASE_URL="http://localhost:8080/api/v1/dashboard"
API_VERSION="v1"

echo "🚀 Iniciando pruebas del endpoint Dashboard"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "API Version: $API_VERSION"
echo ""

# Función para mostrar respuestas de manera legible
show_response() {
    local test_name="$1"
    local response="$2"
    local status_code="$3"
    
    echo "📋 $test_name"
    echo "Status: $status_code"
    echo "Response:"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo "----------------------------------------"
    echo ""
}

# Función para verificar que el servidor esté corriendo
check_server() {
    echo "🔍 Verificando que el servidor esté corriendo..."
    local health_response=$(curl -s -w "%{http_code}" "http://localhost:8080/health")
    local status_code="${health_response: -3}"
    local body="${health_response%???}"
    
    if [ "$status_code" = "200" ]; then
        echo "✅ Servidor está corriendo en puerto 8080"
        echo ""
        return 0
    else
        echo "❌ Servidor no está corriendo en puerto 8080"
        echo "   Status: $status_code"
        echo "   Response: $body"
        echo ""
        echo "💡 Para iniciar el servidor, ejecuta:"
        echo "   make run"
        echo "   # o"
        echo "   go run cmd/api/main.go"
        echo ""
        return 1
    fi
}

# Función para probar endpoint sin filtros
test_no_filters() {
    echo "🧪 Probando endpoint sin filtros..."
    local response=$(curl -s -w "%{http_code}" "$BASE_URL")
    local status_code="${response: -3}"
    local body="${response%???}"
    
    show_response "Sin filtros" "$body" "$status_code"
    
    if [ "$status_code" = "200" ]; then
        echo "✅ Endpoint responde correctamente sin filtros"
    else
        echo "❌ Endpoint falló sin filtros"
    fi
    echo ""
}

# Función para probar filtro por customer_ids
test_customer_filter() {
    echo "🧪 Probando filtro por customer_ids..."
    
    # Probar con un solo customer ID
    echo "   - Con customer_id=1"
    local response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=1")
    local status_code="${response: -3}"
    local body="${response%???}"
    show_response "Filtro customer_ids=1" "$body" "$status_code"
    
    # Probar con múltiples customer IDs
    echo "   - Con customer_ids=1,2,3"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=1,2,3")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro customer_ids=1,2,3" "$body" "$status_code"
    
    # Probar con customer ID inválido
    echo "   - Con customer_id inválido (-1)"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=-1")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro customer_ids=-1 (inválido)" "$body" "$status_code"
    
    echo ""
}

# Función para probar filtro por project_ids
test_project_filter() {
    echo "🧪 Probando filtro por project_ids..."
    
    # Probar con un solo project ID
    echo "   - Con project_id=1"
    local response=$(curl -s -w "%{http_code}" "$BASE_URL?project_ids=1")
    local status_code="${response: -3}"
    local body="${response%???}"
    show_response "Filtro project_ids=1" "$body" "$status_code"
    
    # Probar con múltiples project IDs
    echo "   - Con project_ids=1,2,3"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?project_ids=1,2,3")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro project_ids=1,2,3" "$body" "$status_code"
    
    echo ""
}

# Función para probar filtro por campaign_ids
test_campaign_filter() {
    echo "🧪 Probando filtro por campaign_ids..."
    
    # Probar con un solo campaign ID
    echo "   - Con campaign_id=1"
    local response=$(curl -s -w "%{http_code}" "$BASE_URL?campaign_ids=1")
    local status_code="${response: -3}"
    local body="${response%???}"
    show_response "Filtro campaign_ids=1" "$body" "$status_code"
    
    # Probar con múltiples campaign IDs
    echo "   - Con campaign_ids=1,2,3"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?campaign_ids=1,2,3")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro campaign_ids=1,2,3" "$body" "$status_code"
    
    echo ""
}

# Función para probar filtro por field_ids
test_field_filter() {
    echo "🧪 Probando filtro por field_ids..."
    
    # Probar con un solo field ID
    echo "   - Con field_id=1"
    local response=$(curl -s -w "%{http_code}" "$BASE_URL?field_ids=1")
    local status_code="${response: -3}"
    local body="${response%???}"
    show_response "Filtro field_ids=1" "$body" "$status_code"
    
    # Probar con múltiples field IDs
    echo "   - Con field_ids=1,2,3"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?field_ids=1,2,3")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro field_ids=1,2,3" "$body" "$status_code"
    
    echo ""
}

# Función para probar combinación de filtros
test_combined_filters() {
    echo "🧪 Probando combinación de filtros..."
    
    # Probar con customer_ids y project_ids
    echo "   - Con customer_ids=1,2 y project_ids=10,20"
    local response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=1,2&project_ids=10,20")
    local status_code="${response: -3}"
    local body="${response%???}"
    show_response "Filtro combinado customer_ids=1,2&project_ids=10,20" "$body" "$status_code"
    
    # Probar con todos los filtros
    echo "   - Con todos los filtros: customer_ids=1&project_ids=10&campaign_ids=100&field_ids=1000"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=1&project_ids=10&campaign_ids=100&field_ids=1000")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro combinado completo" "$body" "$status_code"
    
    echo ""
}

# Función para probar casos de error
test_error_cases() {
    echo "🧪 Probando casos de error..."
    
    # Probar con customer_id inválido (string)
    echo "   - Con customer_ids inválido (string)"
    local response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=abc")
    local status_code="${response: -3}"
    local body="${response%???}"
    show_response "Filtro customer_ids=abc (inválido)" "$body" "$status_code"
    
    # Probar con project_id inválido (cero)
    echo "   - Con project_ids inválido (cero)"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?project_ids=0")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro project_ids=0 (inválido)" "$body" "$status_code"
    
    # Probar con formato malformado
    echo "   - Con formato malformado (coma al final)"
    response=$(curl -s -w "%{http_code}" "$BASE_URL?customer_ids=1,")
    status_code="${response: -3}"
    body="${response%???}"
    show_response "Filtro customer_ids=1, (malformado)" "$body" "$status_code"
    
    echo ""
}

# Función para verificar estructura de respuesta
test_response_structure() {
    echo "🧪 Verificando estructura de respuesta..."
    
    local response=$(curl -s "$BASE_URL")
    
    # Verificar que la respuesta sea JSON válido
    if echo "$response" | jq '.' >/dev/null 2>&1; then
        echo "✅ Respuesta es JSON válido"
        
        # Verificar campos principales
        local has_metrics=$(echo "$response" | jq -r '.metrics // empty')
        if [ -n "$has_metrics" ]; then
            echo "✅ Campo 'metrics' presente"
        else
            echo "❌ Campo 'metrics' ausente"
        fi
        
        local has_crop_incidence=$(echo "$response" | jq -r '.crop_incidence // empty')
        if [ -n "$has_crop_incidence" ]; then
            echo "✅ Campo 'crop_incidence' presente"
        else
            echo "❌ Campo 'crop_incidence' ausente"
        fi
        
        local has_management_balance=$(echo "$response" | jq -r '.management_balance // empty')
        if [ -n "$has_management_balance" ]; then
            echo "✅ Campo 'management_balance' presente"
        else
            echo "❌ Campo 'management_balance' presente"
        fi
        
        local has_detailed_management_balance=$(echo "$response" | jq -r '.detailed_management_balance // empty')
        if [ -n "$has_detailed_management_balance" ]; then
            echo "✅ Campo 'detailed_management_balance' presente"
        else
            echo "❌ Campo 'detailed_management_balance' ausente"
        fi
        
    else
        echo "❌ Respuesta no es JSON válido"
    fi
    
    echo ""
}

# Función para mostrar resumen de pruebas
show_summary() {
    echo "📊 Resumen de Pruebas"
    echo "====================="
    echo ""
    echo "✅ Pruebas completadas"
    echo ""
    echo "🔗 Endpoint probado: $BASE_URL"
    echo "📝 Filtros soportados:"
    echo "   - customer_ids (array de int64)"
    echo "   - project_ids (array de int64)"
    echo "   - campaign_ids (array de int64)"
    echo "   - field_ids (array de int64)"
    echo ""
    echo "💡 Ejemplos de uso:"
    echo "   $BASE_URL?customer_ids=1,2,3"
    echo "   $BASE_URL?project_ids=10&campaign_ids=100"
    echo "   $BASE_URL?field_ids=1000,2000,3000"
    echo ""
    echo "🚀 Para ejecutar el servidor:"
    echo "   make run"
    echo "   # o"
    echo "   go run cmd/api/main.go"
    echo ""
}

# Función principal
main() {
    echo "🎯 Dashboard API - Pruebas con curl"
    echo "==================================="
    echo ""
    
    # Verificar que el servidor esté corriendo
    if ! check_server; then
        exit 1
    fi
    
    # Ejecutar todas las pruebas
    test_no_filters
    test_customer_filter
    test_project_filter
    test_campaign_filter
    test_field_filter
    test_combined_filters
    test_error_cases
    test_response_structure
    
    # Mostrar resumen
    show_summary
}

# Ejecutar función principal
main
