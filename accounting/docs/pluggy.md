Client ID: 1fe09f72-ca31-4139-904b-5fa9172767ce
Client Secret: VzBM_pWK1Z7hPtqu_SYfiXJoAhvIS4Cw0xZ1c5kSX1I

claude mcp add --transport http pluggy-docs https://mcp.pluggy.ai/mcp

# Flask - app.py
from pluggy_sdk import PluggyClient
import os

# First, authenticate to get API Key
curl -X POST https://api.pluggy.ai/auth \
  -H "Content-Type: application/json" \
  -d '{"clientId":"YOUR_CLIENT_ID","clientSecret":"YOUR_CLIENT_SECRET"}'

# Then create Connect Token
curl -X POST https://api.pluggy.ai/connect_token \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: YOUR_API_KEY" \
  -d '{"options":{"clientUserId":"user-123"}}'

# Flask - app.py
from pluggy_sdk import PluggyClient
import os

pluggy = PluggyClient(
    client_id=os.environ['CLIENT_ID'],
    client_secret=os.environ['CLIENT_SECRET']
)

@app.route('/api/connect-token', methods=['POST'])
def create_connect_token():
    client_user_id = request.json.get('clientUserId')

    connect_token = pluggy.create_connect_token(
        client_user_id=client_user_id
    )

    return {'accessToken': connect_token.access_token}

{
  "connectToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6IjFmZTA5ZjcyLWNhMzEtNDEzOS05MDRiLTVmYTkxNzI3NjdjZSIsImRhdGEiOiI2YmNhMTBkYTQwZTJhM2I0NDAwYWYwYWU3NDgyM2M2YTozNDg4ODYwMmIwY2I4MzViNDkzNTc1NDUyZjcyMzJkN2U0MDFmMGFkNDdmMjc2ZDNlZDEyYmU4Y2RmOTQ2NGE0NWI5YjJmZjYwYjk4YzI5NjZhNjllYjI4YjRjZjVmMDM3ZDU1MzZiOGRiOWVmNmQ3YWE3MWQ0YWE0NmEyZWRiN2EzZmQ5OGIzMDRlMTY5MDU5ZjhkZmMzM2ZkNDYzNjk2YjE5MDVkMWY1YzY2ZTEyY2U1YjMzYTQzNWZiMDdiN2ViNDg1ODliMjZjZjNiYzFjMjJhOGMzNmZlMjNmNjRmZThiOGRjZTZiMjdkMTU3NDQxMmNkYjM0ZTViMWE2MzNkIiwiaWF0IjoxNzk0MTU1ODU4LCJleHAiOjE3OTQxNTc2NTh9.FSCNp9g9_ygId5WV4GNbZ4NCMzQ1eBpKqX0UYKgv9P8xpEYx3M1lSeyw7iqPTlNNdXuU9w45dBUELJpEJ8MPT1ftSdi9-9MTMeQNjS6D3pa7mLhGVqL8gf0aquSt9ReyjdEp6qZOXJJDnCppVuPig_IHOhVcLxVY7VpRHfVe52Km52Zvvdgr2SgJ7BrFpaeZ-5Wea5HipKKOanOeeafxxc5xQLCaoai0mfCfCcc24P0ysGC_FHjXAyQDa_yUgzmwDFDZIcf7msi9cPyVSUPTRoTXKXCI03DkX-B5PNjcxVUfC-yselxfH4QfRfmMjwmzDWSdV8mCcuDXP470y40i_w"
}


<!-- Add to your HTML -->
<script src="https://cdn.pluggy.ai/connect/v2/pluggy-connect.js"></script>

<script>
// Fetch token from your backend (Step 2)
const connectToken = await fetch('/api/connect-token')
  .then(res => res.json())
  .then(data => data.accessToken)

const pluggyConnect = new window.PluggyConnect({
  connectToken: connectToken,
  includeSandbox: true,
  onSuccess: (itemData) => {
    console.log('Connected!', itemData)
    // Send itemData.item.id to your backend
  },
  onError: (error) => {
    console.error('Connection failed', error)
  }
})

pluggyConnect.init()
</script>