// API Configuration
// Change this to your deployed backend URL

const CONFIG = {
    // Development (local)
    API_URL_DEV: 'http://localhost:8080',
    
    // Production (update this after deploying to Railway)
    API_URL_PROD: 'https://txnflow-production.up.railway.app',
    
    // Auto-detect environment
    get API_URL() {
        // Use production URL if hostname is not localhost
        if (window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
            return this.API_URL_PROD;
        }
        return this.API_URL_DEV;
    },
    
    // Polling intervals (in milliseconds)
    STATS_POLL_INTERVAL: 5000,      // 5 seconds
    TRANSACTIONS_POLL_INTERVAL: 3000, // 3 seconds
    
    // Pagination
    TRANSACTIONS_PER_PAGE: 10,
    
    // Auto-refresh
    AUTO_REFRESH: true
};

// Log current configuration
console.log('🔧 TxnFlow Config:', {
    API_URL: CONFIG.API_URL,
    Environment: window.location.hostname === 'localhost' ? 'Development' : 'Production'
});
