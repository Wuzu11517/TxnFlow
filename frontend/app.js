// TxnFlow Frontend Application
// Main application logic

// State
let currentOffset = 0;
let statsInterval = null;
let transactionsInterval = null;

// DOM Elements
const elements = {
    // Stats
    statTotal: document.getElementById('statTotal'),
    statConfirmed: document.getElementById('statConfirmed'),
    statProcessing: document.getElementById('statProcessing'),
    statError: document.getElementById('statError'),
    
    // Form
    submitForm: document.getElementById('submitForm'),
    txHashInput: document.getElementById('txHash'),
    chainIdSelect: document.getElementById('chainId'),
    submitBtn: document.getElementById('submitBtn'),
    clearBtn: document.getElementById('clearBtn'),
    
    // Transactions
    transactionsBody: document.getElementById('transactionsBody'),
    loadingState: document.getElementById('loadingState'),
    emptyState: document.getElementById('emptyState'),
    transactionsTable: document.getElementById('transactionsTable'),
    refreshBtn: document.getElementById('refreshBtn'),
    loadMoreBtn: document.getElementById('loadMoreBtn'),
    loadMoreContainer: document.getElementById('loadMoreContainer'),
    
    // Modal
    detailModal: document.getElementById('detailModal'),
    detailContent: document.getElementById('detailContent'),
    closeModal: document.getElementById('closeModal'),
    
    // Alerts
    alertContainer: document.getElementById('alertContainer'),
    
    // API Status
    apiStatus: document.getElementById('apiStatus')
};

// ============================================================================
// Utility Functions
// ============================================================================

// Format timestamp to relative time
function timeAgo(timestamp) {
    const now = new Date();
    const past = new Date(timestamp);
    const seconds = Math.floor((now - past) / 1000);
    
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
}

// Truncate hash for display
function truncateHash(hash) {
    if (!hash || hash.length < 15) return hash;
    return `${hash.substring(0, 10)}...${hash.substring(hash.length - 8)}`;
}

// Copy to clipboard
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        showAlert('Copied to clipboard!', 'success');
    } catch (err) {
        console.error('Failed to copy:', err);
        showAlert('Failed to copy', 'error');
    }
}

// Format chain name
function getChainName(chainId) {
    const chains = {
        1: 'Ethereum',
        137: 'Polygon',
        42161: 'Arbitrum'
    };
    return chains[chainId] || `Chain ${chainId}`;
}

// Get status badge HTML
function getStatusBadge(status) {
    const badges = {
        'CONFIRMED': { class: 'confirmed', icon: '✅', text: 'Confirmed' },
        'RECEIVED': { class: 'received', icon: '📥', text: 'Received' },
        'FETCHING': { class: 'fetching', icon: '🔄', text: 'Fetching' },
        'ERROR': { class: 'error', icon: '❌', text: 'Error' }
    };
    
    const badge = badges[status] || { class: 'received', icon: '❓', text: status };
    return `<span class="status-badge ${badge.class}">${badge.icon} ${badge.text}</span>`;
}

// Get Etherscan link
function getEtherscanLink(hash) {
    return `https://etherscan.io/tx/${hash}`;
}

// ============================================================================
// Alert System
// ============================================================================

function showAlert(message, type = 'info', duration = 5000) {
    const alertEl = document.createElement('div');
    alertEl.className = `alert-enter max-w-sm w-full bg-white shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden`;
    
    const colors = {
        success: { bg: 'bg-green-50', border: 'border-green-400', text: 'text-green-800', icon: '✅' },
        error: { bg: 'bg-red-50', border: 'border-red-400', text: 'text-red-800', icon: '❌' },
        info: { bg: 'bg-blue-50', border: 'border-blue-400', text: 'text-blue-800', icon: 'ℹ️' },
        warning: { bg: 'bg-yellow-50', border: 'border-yellow-400', text: 'text-yellow-800', icon: '⚠️' }
    };
    
    const color = colors[type] || colors.info;
    
    alertEl.innerHTML = `
        <div class="p-4 ${color.bg} border-l-4 ${color.border}">
            <div class="flex items-center">
                <span class="text-2xl mr-3">${color.icon}</span>
                <p class="${color.text} font-medium">${message}</p>
            </div>
        </div>
    `;
    
    elements.alertContainer.appendChild(alertEl);
    
    // Auto-remove
    setTimeout(() => {
        alertEl.classList.add('alert-exit');
        setTimeout(() => alertEl.remove(), 300);
    }, duration);
}

// ============================================================================
// API Functions
// ============================================================================

// Check API status
async function checkAPIStatus() {
    try {
        const response = await fetch(`${CONFIG.API_URL}/stats`);
        if (response.ok) {
            elements.apiStatus.innerHTML = `
                <div class="w-2 h-2 rounded-full bg-green-500 status-dot"></div>
                <span class="text-green-600 font-medium">Connected</span>
            `;
            return true;
        }
    } catch (error) {
        elements.apiStatus.innerHTML = `
            <div class="w-2 h-2 rounded-full bg-red-500"></div>
            <span class="text-red-600 font-medium">Disconnected</span>
        `;
        return false;
    }
}

// Fetch stats
async function fetchStats() {
    try {
        const response = await fetch(`${CONFIG.API_URL}/stats`);
        if (!response.ok) throw new Error('Failed to fetch stats');
        
        const data = await response.json();
        updateStats(data);
    } catch (error) {
        console.error('Error fetching stats:', error);
    }
}

// Update stats display
function updateStats(data) {
    elements.statTotal.textContent = data.total || 0;
    elements.statConfirmed.textContent = data.by_status?.CONFIRMED || 0;
    
    // Calculate processing (RECEIVED + FETCHING)
    const processing = (data.by_status?.RECEIVED || 0) + (data.by_status?.FETCHING || 0);
    elements.statProcessing.textContent = processing;
    
    elements.statError.textContent = data.by_status?.ERROR || 0;
}

// Submit transaction
async function submitTransaction(hash, chainId) {
    try {
        elements.submitBtn.disabled = true;
        elements.submitBtn.innerHTML = `
            <svg class="animate-spin h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            Submitting...
        `;
        
        const response = await fetch(`${CONFIG.API_URL}/transactions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                transaction_hash: hash,
                chain_id: parseInt(chainId)
            })
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || 'Failed to submit transaction');
        }
        
        const data = await response.json();
        
        showAlert(`Transaction submitted! Status: ${data.status}`, 'success');
        
        // Clear form
        elements.txHashInput.value = '';
        
        // Refresh data
        setTimeout(() => {
            fetchTransactions();
            fetchStats();
        }, 500);
        
    } catch (error) {
        console.error('Error submitting transaction:', error);
        showAlert(`Error: ${error.message}`, 'error');
    } finally {
        elements.submitBtn.disabled = false;
        elements.submitBtn.innerHTML = `
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
            </svg>
            Submit Transaction
        `;
    }
}

// Fetch transactions
async function fetchTransactions(offset = 0, append = false) {
    try {
        if (!append) {
            elements.loadingState.classList.remove('hidden');
            elements.transactionsTable.classList.add('hidden');
            elements.emptyState.classList.add('hidden');
        }
        
        const response = await fetch(
            `${CONFIG.API_URL}/transactions?limit=${CONFIG.TRANSACTIONS_PER_PAGE}&offset=${offset}`
        );
        
        if (!response.ok) throw new Error('Failed to fetch transactions');
        
        const data = await response.json();
        
        elements.loadingState.classList.add('hidden');
        
        if (!data.data || data.data.length === 0) {
            if (offset === 0) {
                elements.emptyState.classList.remove('hidden');
                elements.transactionsTable.classList.add('hidden');
            }
            elements.loadMoreContainer.classList.add('hidden');
            return;
        }
        
        elements.emptyState.classList.add('hidden');
        elements.transactionsTable.classList.remove('hidden');
        
        if (append) {
            appendTransactions(data.data);
        } else {
            renderTransactions(data.data);
        }
        
        // Show/hide load more button
        if (data.count === CONFIG.TRANSACTIONS_PER_PAGE) {
            elements.loadMoreContainer.classList.remove('hidden');
        } else {
            elements.loadMoreContainer.classList.add('hidden');
        }
        
    } catch (error) {
        console.error('Error fetching transactions:', error);
        elements.loadingState.classList.add('hidden');
        showAlert('Failed to load transactions', 'error');
    }
}

// Render transactions
function renderTransactions(transactions) {
    elements.transactionsBody.innerHTML = transactions.map(tx => createTransactionRow(tx)).join('');
}

// Append transactions
function appendTransactions(transactions) {
    elements.transactionsBody.innerHTML += transactions.map(tx => createTransactionRow(tx)).join('');
}

// Create transaction row
function createTransactionRow(tx) {
    return `
        <tr class="hover:bg-gray-50 cursor-pointer" onclick="viewTransactionDetails('${tx.transaction_hash}')">
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center gap-2">
                    <span class="hash-display font-mono">${truncateHash(tx.transaction_hash)}</span>
                    <button onclick="event.stopPropagation(); copyToClipboard('${tx.transaction_hash}')" 
                            class="copy-btn text-gray-400 hover:text-gray-600"
                            title="Copy full hash">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                        </svg>
                    </button>
                </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                ${getChainName(tx.chain_id)}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
                ${getStatusBadge(tx.status)}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                ${timeAgo(tx.created_at)}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm">
                <button onclick="event.stopPropagation(); viewTransactionDetails('${tx.transaction_hash}')"
                        class="text-blue-600 hover:text-blue-800 font-medium">
                    View Details →
                </button>
            </td>
        </tr>
    `;
}

// View transaction details
async function viewTransactionDetails(hash) {
    try {
        const response = await fetch(`${CONFIG.API_URL}/transactions/${hash}`);
        if (!response.ok) throw new Error('Failed to fetch transaction details');
        
        const tx = await response.json();
        
        elements.detailContent.innerHTML = `
            <div class="space-y-6">
                <!-- Header -->
                <div class="flex items-start justify-between">
                    <div>
                        <h4 class="text-lg font-semibold text-gray-900">Transaction Hash</h4>
                        <p class="font-mono text-sm text-gray-600 mt-1 break-all">${tx.transaction_hash}</p>
                    </div>
                    ${getStatusBadge(tx.status)}
                </div>

                <!-- Basic Info -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div class="bg-gray-50 p-4 rounded-lg">
                        <p class="text-sm text-gray-600 font-medium">Chain</p>
                        <p class="text-lg font-semibold mt-1">${getChainName(tx.chain_id)}</p>
                    </div>
                    <div class="bg-gray-50 p-4 rounded-lg">
                        <p class="text-sm text-gray-600 font-medium">Created</p>
                        <p class="text-lg font-semibold mt-1">${new Date(tx.created_at).toLocaleString()}</p>
                    </div>
                </div>

                ${tx.from_address ? `
                <!-- Blockchain Data -->
                <div class="border-t pt-4">
                    <h5 class="font-semibold text-gray-900 mb-3">Blockchain Data</h5>
                    <div class="space-y-3">
                        <div>
                            <p class="text-sm text-gray-600">From Address</p>
                            <div class="flex items-center gap-2 mt-1">
                                <p class="font-mono text-sm text-gray-900 break-all">${tx.from_address}</p>
                                <button onclick="copyToClipboard('${tx.from_address}')" class="text-gray-400 hover:text-gray-600">
                                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                                    </svg>
                                </button>
                            </div>
                        </div>
                        <div>
                            <p class="text-sm text-gray-600">To Address</p>
                            <div class="flex items-center gap-2 mt-1">
                                <p class="font-mono text-sm text-gray-900 break-all">${tx.to_address}</p>
                                <button onclick="copyToClipboard('${tx.to_address}')" class="text-gray-400 hover:text-gray-600">
                                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                                    </svg>
                                </button>
                            </div>
                        </div>
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <div>
                                <p class="text-sm text-gray-600">Value</p>
                                <p class="font-mono text-sm font-semibold mt-1">${tx.value || '0'} wei</p>
                            </div>
                            <div>
                                <p class="text-sm text-gray-600">Block Number</p>
                                <p class="font-mono text-sm font-semibold mt-1">${tx.block_number || 'N/A'}</p>
                            </div>
                            <div>
                                <p class="text-sm text-gray-600">Gas Used</p>
                                <p class="font-mono text-sm font-semibold mt-1">${tx.gas_used || 'N/A'}</p>
                            </div>
                        </div>
                    </div>
                </div>
                ` : ''}

                ${tx.error_reason ? `
                <!-- Error Info -->
                <div class="bg-red-50 border border-red-200 rounded-lg p-4">
                    <h5 class="font-semibold text-red-900 mb-2">Error Details</h5>
                    <p class="text-sm text-red-800">${tx.error_reason}</p>
                </div>
                ` : ''}

                <!-- Actions -->
                <div class="flex gap-3 pt-4 border-t">
                    <a href="${getEtherscanLink(tx.transaction_hash)}" 
                       target="_blank"
                       class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-center font-medium">
                        View on Etherscan →
                    </a>
                    <button onclick="copyToClipboard('${tx.transaction_hash}')"
                            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 font-medium">
                        Copy Hash
                    </button>
                </div>
            </div>
        `;
        
        elements.detailModal.classList.remove('hidden');
        elements.detailModal.classList.add('flex');
        
    } catch (error) {
        console.error('Error fetching transaction details:', error);
        showAlert('Failed to load transaction details', 'error');
    }
}

// Make globally available
window.viewTransactionDetails = viewTransactionDetails;
window.copyToClipboard = copyToClipboard;

// ============================================================================
// Event Handlers
// ============================================================================

// Submit form
elements.submitForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const hash = elements.txHashInput.value.trim();
    const chainId = elements.chainIdSelect.value;
    
    if (!hash) {
        showAlert('Please enter a transaction hash', 'warning');
        return;
    }
    
    if (!hash.startsWith('0x')) {
        showAlert('Transaction hash must start with 0x', 'warning');
        return;
    }
    
    submitTransaction(hash, chainId);
});

// Clear form
elements.clearBtn.addEventListener('click', () => {
    elements.txHashInput.value = '';
});

// Demo transaction buttons
document.querySelectorAll('.demo-tx').forEach(btn => {
    btn.addEventListener('click', () => {
        const hash = btn.dataset.hash;
        elements.txHashInput.value = hash;
        elements.txHashInput.focus();
    });
});

// Refresh transactions
elements.refreshBtn.addEventListener('click', () => {
    currentOffset = 0;
    fetchTransactions();
    fetchStats();
});

// Load more
elements.loadMoreBtn.addEventListener('click', () => {
    currentOffset += CONFIG.TRANSACTIONS_PER_PAGE;
    fetchTransactions(currentOffset, true);
});

// Close modal
elements.closeModal.addEventListener('click', () => {
    elements.detailModal.classList.add('hidden');
    elements.detailModal.classList.remove('flex');
});

// Close modal on background click
elements.detailModal.addEventListener('click', (e) => {
    if (e.target === elements.detailModal) {
        elements.detailModal.classList.add('hidden');
        elements.detailModal.classList.remove('flex');
    }
});

// ============================================================================
// Auto-refresh
// ============================================================================

function startAutoRefresh() {
    if (!CONFIG.AUTO_REFRESH) return;
    
    // Stats refresh
    statsInterval = setInterval(fetchStats, CONFIG.STATS_POLL_INTERVAL);
    
    // Transactions refresh (only if not on subsequent pages)
    transactionsInterval = setInterval(() => {
        if (currentOffset === 0) {
            fetchTransactions();
        }
    }, CONFIG.TRANSACTIONS_POLL_INTERVAL);
}

function stopAutoRefresh() {
    if (statsInterval) clearInterval(statsInterval);
    if (transactionsInterval) clearInterval(transactionsInterval);
}

// ============================================================================
// Initialization
// ============================================================================

async function init() {
    console.log('🚀 TxnFlow Frontend Initialized');
    console.log('📡 API URL:', CONFIG.API_URL);
    
    // Check API status
    const apiOnline = await checkAPIStatus();
    
    if (!apiOnline) {
        showAlert('Warning: Cannot connect to API. Make sure the backend is running.', 'warning', 10000);
    }
    
    // Initial data load
    await fetchStats();
    await fetchTransactions();
    
    // Start auto-refresh
    startAutoRefresh();
    
    console.log('✅ Frontend ready!');
}

// Start app when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}

// Cleanup on page unload
window.addEventListener('beforeunload', stopAutoRefresh);
