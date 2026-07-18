import ThunderAPI from 'thunderapi';

const api = new ThunderAPI();

try {
  const player = await api.getPlayer('Dark#598');
  console.log(JSON.stringify(player, null, 2));
} catch (err) {
  console.error('Error:', err.message);
  if (err.response) console.error('Status:', err.response.status);
}
