import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Нагрузка задаётся через open model (constant-arrival-rate),
// чтобы гарантировать фиксированный поток 1000 RPS независимо от времени ответа.
//
// rate: 1000, 
// timeUnit: "1s" → 1000 итераций в секунду (1 итерация = 1 HTTP-запрос).
//
// preAllocatedVUs рассчитаны по формуле:
// VUs ≈ RPS × response_time.
//
// При SLI latency ≤100ms (0.1s):
// 1000 × 0.1 = 100 VUs (теоретический минимум).
// С учётом jitter, p99 и накладных расходов клиента берём запас → 200 VUs.
//
// maxVUs = 300 — резерв на случай роста latency,
// чтобы избежать dropped_iterations.
//
// duration 2m — достаточная выборка для проверки SLI 99.99%.

const BASE_URL = 'http://app:8080';

export const options = {
    scenarios: {
        auth_scenario: {
            executor: 'constant-arrival-rate',
            rate: 100,           // 10% от 1000
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 100,
            maxVUs: 150,
            exec: 'authScenario',
        },

        pvz_scenario: {
            executor: 'constant-arrival-rate',
            rate: 450,           // 45% от 1000
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 450,
            maxVUs: 550,
            exec: 'pvzScenario',
        },

        reception_scenario: {
            executor: 'constant-arrival-rate',
            rate: 450,           // 45% от 1000
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 450,
            maxVUs: 550,
            exec: 'receptionE2EScenario',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.0001'],
        http_req_duration: ['p(99)<100'],
    },
};

function getAuthHeaders(token) {
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
    };
}

export function setup() {
    function registerAndLogin(role) {
        const res = http.post(`${BASE_URL}/dummyLogin`, JSON.stringify({ role }),
            { headers: { "Content-Type": "text/plain; charset=utf-8" } }
        );
        check(res, { [`dummyLogin ${role}`]: (r) => r.status === 200 });

        if (!res.body || res.body === "") {
            fail(`dummyLogin failed for role: ${role}, status: ${res.status}, res: ${res.body}`);
        }

        return res.body;
    }

    const employeeToken = registerAndLogin('employee');

    const moderatorToken = registerAndLogin('moderator');

    const moderatorHeaders = getAuthHeaders(moderatorToken); // ← определяем до использования

    for (let i = 0; i < 10; i++) {
        http.post(`${BASE_URL}/pvz`, JSON.stringify({
            city: "Москва",
            id: crypto.randomUUID(),
            registrationDate: "2025-09-22T18:04:04.605Z",
        }), { headers: moderatorHeaders });
    }

    for (let i = 0; i < 10; i++) {
        http.post(`${BASE_URL}/pvz`, JSON.stringify({
            city: "Москва",
            id: crypto.randomUUID(),
            registrationDate: "2025-09-22T18:04:04.605Z",
        }), { headers: moderatorHeaders });
    }

    return { employeeToken, moderatorToken };
}


export function authScenario() {
    const email = `load_test_${randomString(8)}@test.com`;
    const password = randomString(6);

    const regRes = http.post(`${BASE_URL}/register`, JSON.stringify({
        "email": email,
        "password": password,
        "role": "employee"
    }));
    const checkReg = check(regRes, { 'register ok': (r) => r.status === 201 });
    if (!checkReg) {
        fail(`register failed, status: ${regRes.status}, res: ${regRes.body}`);
    }

    const loginRes = http.post(`${BASE_URL}/login`, JSON.stringify({
        "email": email,
        "password": password,
    }));
    const checkLogin = check(loginRes, { 'login ok': (r) => r.status === 200 });
    if (!checkLogin) {
        fail(`login failed, status: ${regRes.status}, res: ${regRes.body}`);
    }

    sleep(1);
}

function buildQueryString(params) {
    return Object.entries(params)
        .map(([key, value]) => `${key}=${value}`)
        .join("&");
}

export function pvzScenario(data) {
    const employeeHeader = getAuthHeaders(data.employeeToken);

    const params = buildQueryString({
        startDate: "2008-01-01T00:00:00Z",
        endDate: "2035-01-01T00:00:00Z",
        page: 1,
        limit: 10,
    });

    const pvzRes = http.get(`${BASE_URL}/pvz?${params}`, { headers: employeeHeader });
    const checkResult = check(pvzRes, { 'get pvz ok': (r) => r.status === 200 });

    if (!checkResult) {
        fail(`get pvz failed, status: ${pvzRes.status}, res: ${pvzRes.body}`);
    }

    const pvzData = pvzRes.json();
    if (!Array.isArray(pvzData) || pvzData.length === 0 || !pvzData.some(item => item.pvz && item.pvz.id)) {
        fail("No PVZ found!");
    }

    sleep(1);
}

export function receptionE2EScenario(data) {
    const moderatorHeaders = getAuthHeaders(data.moderatorToken);
    const employeeHeaders = getAuthHeaders(data.employeeToken);

    const PVZData = {
        "city": "Москва",
        "id": crypto.randomUUID(),
        "registrationDate": "2025-09-22T18:04:04.605Z",
    }

    // 1. Создание PVZ
    const pvzRes = postJSON(`${BASE_URL}/pvz`, PVZData, moderatorHeaders, 201);
    const pvzId = pvzRes.json().id;

    // 2. Создание приемки
    postJSON(`${BASE_URL}/receptions`, { pvzId }, employeeHeaders, 201);

    // 3. Добавление продукта
    postJSON(`${BASE_URL}/products`, { pvzId, type: "одежда" }, employeeHeaders, 201);

    // 4. Удаление последнего продукта
    postJSON(`${BASE_URL}/pvz/${pvzId}/delete_last_product`, null, employeeHeaders, 200);

    // 5. Закрытие последней приемки
    postJSON(`${BASE_URL}/pvz/${pvzId}/close_last_reception`, null, employeeHeaders, 200);

    sleep(1);
}

function postJSON(url, body, headers, expectedStatus) {
    const res = http.post(url, body ? JSON.stringify(body) : null, { headers });
    if (res.status !== expectedStatus) {
        console.error(`❌ ${url} failed, status: ${res.status}, body: ${res.body}`);
        fail(`Request failed`);
    }
    return res;
}