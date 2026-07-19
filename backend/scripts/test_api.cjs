const ThunderAPI = require('thunderapi').ThunderAPI;
const api = new ThunderAPI();

async function main() {
  try {
    console.log('Fetching player Dark#598...');
    const player = await api.getPlayer('Dark#598', false);
    console.log(JSON.stringify(player, null, 2));
  } catch (err) {
    console.error('Error:', err.message);
    if (err.response) console.error('Status:', err.response.status);
    if (err.stack) console.error(err.stack.split('\n').slice(0, 5).join('\n'));
  }
}

main();
