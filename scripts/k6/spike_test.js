import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import { vu } from 'k6/execution';

const reserveSuccess = new Counter('ticket_reserve_200');
const checkoutSuccess = new Counter('ticket_checkout_200');
const reserveFail = new Counter('ticket_reserve_fail');

const users = new SharedArray('users', function () {
  return JSON.parse(open('./users.json'));
});

export const options = {
  scenarios: {
    instant_rush: {
      executor: 'per-vu-iterations',
      vus: 5000,
      iterations: 1,
      maxDuration: '2m', 
    },
  },
};

export default function () {
  const currentUser = users[(vu.idInTest - 1) % users.length];

  const payload = JSON.stringify({
    event_id: 'concert_2026',
    ticket_type: 'VIP',
    user_name: currentUser.user_name,
    phone_num: currentUser.phone_num,
    area: 1
  });

  const params = { headers: { 'Content-Type': 'application/json' }, timeout: '10s' };

  sleep(Math.random() * 0.05); 

  // 【階段一：搶佔保留位】
  const resReserve = http.post('http://localhost:8080/api/tickets/reserve', payload, params);

  if (resReserve.status === 200) {
    reserveSuccess.add(1);
    
    // 模擬真實使用者看到畫面後，花 1~2 秒點擊結帳
    sleep(Math.random() * 1 + 1);

    // 【階段二：結帳扣款】
    const resCheckout = http.post('http://localhost:8080/api/tickets/checkout', payload, params);
    
    check(resCheckout, {
      '結帳成功 (200)': (r) => r.status === 200,
    });
    
    if (resCheckout.status === 200) {
      checkoutSuccess.add(1);
    }
  } else {
    reserveFail.add(1);
  }
}