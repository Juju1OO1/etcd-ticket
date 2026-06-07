import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import { vu } from 'k6/execution';

// --- 1. 自訂監控指標 ---

const broadcastLatency = new Trend('ws_broadcast_latency');
const wsConnections = new Counter('ws_connections_success');
const wsErrors = new Counter('ws_errors');
const messagesReceived = new Counter('ws_messages_received');

// --- 2. 載入測試資料 ---
const users = new SharedArray('users with tokens', function () {
  return JSON.parse(open('./users.json'));
});

// --- 3. 壓測情境設定 ---
export const options = {
  scenarios: {
    websocket_listeners: {
      executor: 'ramping-vus', 
      startVUs: 0,
      stages: [
        { duration: '20s', target: 3000 }, 
        { duration: '1m', target: 3000 },  
        { duration: '10s', target: 0 },   
      ],
    },
  },
};

export default function () {
  const userIndex = vu.idInTest - 1;
  const currentUser = users[userIndex % users.length]; 

  const url = `ws://127.0.0.1:8888/ws`;

  const params = { headers: {} };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', function () {
      wsConnections.add(1);
    });

    socket.on('message', function (msg) {
      messagesReceived.add(1);
      
      try {
        const data = JSON.parse(msg);
        
        if (data.type === 'ticket_available' || data.type === 'ticket_sold') {
          
          if (data.server_timestamp) {
            const currentClientTime = Date.now();
            const latency = currentClientTime - data.server_timestamp;
            broadcastLatency.add(latency);
          }
        }
      } catch (e) {
        console.error('Failed to parse WS message:', msg);
      }
    });

    socket.on('error', function (e) {
      if (e.error() != 'websocket: close sent') {
        wsErrors.add(1);
      }
    });

    socket.setTimeout(function () {
      socket.close();
    }, 90000); 
  });

  check(res, {
    'WebSocket 連線成功 (101)': (r) => r && r.status === 101,
  });
}