# TxnFlow Frontend

Modern, responsive web interface for the TxnFlow blockchain transaction pipeline.

## 🎨 Features

- ✅ **Real-time Stats Dashboard** - Live transaction counts with auto-refresh
- ✅ **Transaction Submission** - Submit Ethereum transactions via form
- ✅ **Transaction List** - Browse recent transactions with status updates
- ✅ **Transaction Details** - View complete transaction information
- ✅ **Famous Transactions** - Pre-loaded examples (first ETH tx, etc.)
- ✅ **Etherscan Integration** - Direct links to verify on Etherscan
- ✅ **Auto-refresh** - Live updates every 3-5 seconds
- ✅ **Responsive Design** - Works on desktop, tablet, and mobile
- ✅ **Copy to Clipboard** - One-click hash copying
- ✅ **Error Handling** - User-friendly error messages

## 🚀 Quick Start

### Option 1: Local Development

```bash
# 1. Make sure backend is running
cd ../
docker-compose up -d

# 2. Open frontend
cd frontend
open index.html
# Or just double-click index.html in your file browser
```

The frontend will automatically connect to `http://localhost:8080`

### Option 2: Deploy to Vercel (Free)

```bash
# 1. Install Vercel CLI
npm install -g vercel

# 2. Deploy
cd frontend
vercel

# Follow prompts - takes 2 minutes!
# You'll get a URL like: https://txnflow.vercel.app
```

### Option 3: Deploy to Netlify (Free)

```bash
# Just drag the frontend/ folder to netlify.com/drop
# Or connect via GitHub for auto-deploy
```

## 📁 File Structure

```
frontend/
├── index.html     # Main page (complete UI)
├── app.js         # Application logic
├── config.js      # API configuration
├── styles.css     # Custom styles
└── README.md      # This file
```

## ⚙️ Configuration

### Update API URL for Production

Edit `config.js`:

```javascript
const CONFIG = {
    // Update this after deploying backend to Railway
    API_URL_PROD: 'https://your-backend.up.railway.app',
    
    // Other settings...
};
```

The frontend automatically detects if you're on localhost (dev) or deployed (prod).

## 🎯 How It Works

### 1. Stats Dashboard
- Polls `/stats` endpoint every 5 seconds
- Shows: Total, Confirmed, Processing, Error counts
- Live updates with visual feedback

### 2. Submit Transaction
- User enters Ethereum transaction hash
- POST to `/transactions`
- Shows success/error alerts
- Auto-refreshes transaction list

### 3. Transaction List
- Fetches from `/transactions?limit=10&offset=0`
- Polls every 3 seconds for updates
- Click row to view details
- "Load More" for pagination

### 4. Transaction Details Modal
- Fetches from `/transactions/{hash}`
- Shows: from, to, value, block, gas
- Links to Etherscan for verification
- Copy buttons for all hashes/addresses

## 🔧 Customization

### Change Refresh Intervals

Edit `config.js`:

```javascript
const CONFIG = {
    STATS_POLL_INTERVAL: 5000,      // Stats refresh (ms)
    TRANSACTIONS_POLL_INTERVAL: 3000, // Transactions refresh (ms)
    AUTO_REFRESH: true              // Enable/disable
};
```

### Change Transactions Per Page

Edit `config.js`:

```javascript
const CONFIG = {
    TRANSACTIONS_PER_PAGE: 10  // Change to 20, 50, etc.
};
```

### Disable Auto-Refresh

Edit `config.js`:

```javascript
const CONFIG = {
    AUTO_REFRESH: false
};
```

## 🎨 Future UI Enhancements

The current frontend is **fully functional** with a clean, professional look. Here are ideas for making it even cooler later:

### Phase 1 Enhancements (Easy - 1-2 hours)
- [ ] Dark mode toggle
- [ ] Animated status transitions
- [ ] Chart.js graphs (transactions over time)
- [ ] Filtering by status/chain
- [ ] Search by hash

### Phase 2 Enhancements (Medium - 3-4 hours)
- [ ] WebSocket real-time updates (no polling)
- [ ] Transaction timeline visualization
- [ ] Gas price trends chart
- [ ] Export to CSV
- [ ] Advanced filters panel

### Phase 3 Enhancements (Advanced - 5+ hours)
- [ ] React/Vue rewrite
- [ ] D3.js network graph (address relationships)
- [ ] Real-time notifications (browser push)
- [ ] Multi-language support
- [ ] PWA (installable app)

## 📱 Browser Support

- ✅ Chrome/Edge (latest)
- ✅ Firefox (latest)
- ✅ Safari (latest)
- ✅ Mobile browsers (iOS Safari, Chrome Mobile)

## 🐛 Troubleshooting

### "Cannot connect to API"

**Problem**: Red "Disconnected" indicator

**Solutions**:
1. Check backend is running: `docker-compose ps`
2. Verify API URL in `config.js`
3. Check browser console for CORS errors
4. Try: `docker-compose restart api`

### CORS Errors

**Problem**: Browser blocks requests due to CORS

**Solution**: Add CORS headers to backend (already configured in Go API)

### Stats Not Updating

**Problem**: Stats stuck at old values

**Solutions**:
1. Check browser console for errors
2. Verify backend is running
3. Hard refresh: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)

### Transactions Not Loading

**Problem**: Empty state or loading forever

**Solutions**:
1. Check backend logs: `docker-compose logs api`
2. Test API directly: `curl http://localhost:8080/transactions`
3. Check database has data: `curl http://localhost:8080/stats`

## 🚀 Deployment Guide

### Deploy Backend First (Railway)

```bash
# 1. Push to GitHub
git push origin main

# 2. Go to railway.app
# 3. New Project → Deploy from GitHub
# 4. Add PostgreSQL database
# 5. Set INFURA_API_KEY
# 6. Copy your URL (e.g., https://txnflow-production.up.railway.app)
```

### Deploy Frontend (Vercel)

```bash
# 1. Update config.js with your Railway URL
# 2. Push to GitHub
# 3. Go to vercel.com
# 4. New Project → Import from GitHub
# 5. Root directory: ./frontend
# 6. Deploy!

# You'll get: https://txnflow.vercel.app
```

### Test Your Deployment

```bash
# 1. Open your Vercel URL
# 2. Submit a real Ethereum transaction:
#    0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060
# 3. Watch it process in real-time
# 4. Click "View Details"
# 5. Click "View on Etherscan" to verify
```

## 📊 Performance

- **Initial Load**: ~1-2 seconds
- **API Calls**: <100ms (localhost), <500ms (deployed)
- **Auto-refresh**: Minimal impact (only updates changed data)
- **Mobile**: Fully responsive, optimized for touch

## 🔐 Security Notes

- ✅ No sensitive data in frontend
- ✅ All API calls go through backend
- ✅ No API keys exposed
- ✅ Input validation on forms
- ✅ XSS protection (no innerHTML with user data)

## 📝 License

Same as TxnFlow backend (MIT)

## 🙋 Support

- **Issues**: Open GitHub issue
- **Questions**: Check backend README
- **Demo**: See live demo at [your-vercel-url]

---

Built with vanilla JavaScript, Tailwind CSS, and ❤️
