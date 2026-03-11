//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Первым делом создает новый ПВЗ
// * Добавляет новую приёмку заказов
// * Добавляет 50 товаров в рамках текущей приёмки заказов
// * Закрывает приёмку заказов

func TestReceptionFullE2E(t *testing.T) {
	moderatorToken := getToken(t, "moderator")
	employeeToken := getToken(t, "employee")

	// 1. Создание PVZ
	pvzID := uuid.New().String()
	pvzRes := post(t, "/pvz", map[string]string{
		"city":             "Москва",
		"id":               pvzID,
		"registrationDate": "1996-09-22T18:04:04.605Z",
	}, moderatorToken)
	require.Equal(t, http.StatusCreated, pvzRes.StatusCode)

	var pvz struct {
		ID               string `json:"id"`
		City             string `json:"city"`
		RegistrationDate string `json:"registrationDate"`
	}
	require.NoError(t, json.NewDecoder(pvzRes.Body).Decode(&pvz))
	require.NotEmpty(t, pvz.ID)
	require.Equal(t, pvzID, pvz.ID)

	// БД: PVZ создан
	assertPVZExists(t, pvz.ID)

	// 2. Открытие приёмки
	recRes := post(t, "/receptions", map[string]string{"pvzId": pvz.ID}, employeeToken)
	require.Equal(t, http.StatusCreated, recRes.StatusCode)

	// БД: приёмка открыта
	assertReceptionStatus(t, pvz.ID, "in_progress")

	// 3. Добавление 50 товаров
	productTypes := []string{"одежда", "электроника", "обувь"}
	for i := range 50 {
		res := post(t, "/products", map[string]string{
			"pvzId": pvz.ID,
			"type":  productTypes[i%len(productTypes)],
		}, employeeToken)
		require.Equal(t, http.StatusCreated, res.StatusCode, "товар %d не добавлен", i+1)
	}

	// БД: ровно 50 товаров
	assertProductCount(t, pvz.ID, 50)

	// 4. Закрытие приёмки
	closeRes := post(t, fmt.Sprintf("/pvz/%s/close_last_reception", pvz.ID), nil, employeeToken)
	require.Equal(t, http.StatusOK, closeRes.StatusCode)

	// БД: приёмка закрыта, товары на месте
	assertReceptionStatus(t, pvz.ID, "close")
	assertProductCount(t, pvz.ID, 50)
}

func post(t *testing.T, path string, body any, token string) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(http.MethodPost, testApp.Server.URL+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	t.Logf("POST %s → %d | body: %s", path, resp.StatusCode, string(bodyBytes))

	return resp
}

func getToken(t *testing.T, role string) string {
	t.Helper()
	resp := post(t, "/dummyLogin", map[string]string{"role": role}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	token := strings.TrimSpace(string(body))
	require.NotEmpty(t, token)
	return token
}

func assertPVZExists(t *testing.T, pvzID string) {
	t.Helper()
	ctx := context.Background()

	var count int
	err := testApp.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM pvz WHERE id = $1`, pvzID,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "pvz %s не найден в БД", pvzID)
}

func assertReceptionStatus(t *testing.T, pvzID string, expectedStatus string) {
	t.Helper()
	ctx := context.Background()

	var status string
	err := testApp.DB.QueryRow(ctx,
		`SELECT  rs.name 
FROM receptions 
JOIN reception_statuses rs ON receptions.status_id = rs.id
WHERE pvz_id = $1 
  AND rs.name = $2
LIMIT 1`,
		pvzID, expectedStatus,
	).Scan(&status)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, status, "неожиданный статус приёмки для pvz %s", pvzID)
}

func assertProductCount(t *testing.T, pvzID string, expected int) {
	t.Helper()
	ctx := context.Background()

	var count int
	err := testApp.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM products p
		JOIN receptions r ON r.id = p.reception_id
		WHERE r.pvz_id = $1
	`, pvzID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, expected, count, "неожиданное кол-во товаров для pvz %s", pvzID)
}
