const fs = require('fs');
const users = [];

for (let i = 1; i <= 5000; i++) {
  users.push({
    user_name: `user_${i}`,
    phone_num: `0900${String(i).padStart(6, '0')}` 
  });
}

fs.writeFileSync(__dirname + '/users.json', JSON.stringify(users, null, 2));

console.log('成功生成 5000 筆測試資料至 users.json');