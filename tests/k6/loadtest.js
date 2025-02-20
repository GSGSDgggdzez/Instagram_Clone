import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        exact_users: {
            executor: 'per-vu-iterations',
            vus: 105,
            iterations: 105,
            maxDuration: '8m',
            gracefulStop: '5s'
        }
    },
    thresholds: {
        'http_req_duration': ['p(95)<2000'],
        'http_req_failed': ['rate<0.01']
    }
};

// Track successful registrations
let successfulRegistrations = 0;
const BASE_URL = 'http://localhost:8090/api/v1';

export default function () {
    const timestamp = Date.now();
    const randomNum = Math.floor(Math.random() * 10000);
    
   // REGISTRATION ASSAULT
   const registerData = {
    username: `user_${timestamp}_${randomNum}`,
    name: "Test User",
    email: `test_${timestamp}_${randomNum}@test.com`,
    password: "test123456",
    phone: `+1${timestamp}`.substring(0, 13),
    language: "en"
};

const registerRes = http.post(`${BASE_URL}/auth/register`, registerData, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    tags: { name: 'RegisterEndpoint' },
    timeout: '10s',
});

if (registerRes.status === 201 || registerRes.status === 200) {
    successfulRegistrations++;
}

    check(registerRes, {
        'registration successful': (r) => r.status === 201 || r.status === 200,
    'response time OK': (r) => r.timings.duration < 2000,
    'response has data': (r) => r.json().data !== undefined,
    'content-type is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
    'response size < 10KB': (r) => r.body.length < 10000
    });

    // LOGIN BOMBARDMENT
    if (registerRes.status === 201 || registerRes.status === 200) {
        const loginData = {
            email: registerData.email,
            password: registerData.password
        };
        
        const loginRes = http.post(`${BASE_URL}/auth/login`, loginData, {
            headers: { 'Content-Type': 'application/json' },
            tags: { name: 'LoginEndpoint' }
        });
        
        check(loginRes, {
            'status is 200': (r) => r.status === 200,
    'response time OK': (r) => r.timings.duration < 2000,
    'response has data': (r) => r.json().data !== undefined,
    'content-type is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
    'response size < 10KB': (r) => r.body.length < 10000
        });
    }

    // Random sleep between 0.5 and 2 seconds to simulate real user behavior
    sleep(Math.random() * 0.2); // Add small random delay between requests
}
