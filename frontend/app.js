const QueryLabApp = {
    sessionID: null,
    isLoading: false,
    
    async init() {
        console.log('QueryLab app initializing...');
        
        // Initialize session
        await this.initializeSession();
        
        // Set up event listeners
        this.setupEventListeners();
        
        // Optional: Setup beforeunload handler
        this.setupUnloadHandler();
    },
    
    async initializeSession() {
        try {
            this.setLoading(true);
            
            const response = await fetch('/api/session', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            const data = await response.json();
            this.sessionID = data.session_id;
            
            console.log('Session initialized:', this.sessionID);
            this.showMessage('Session ready', 'success');
            
        } catch (error) {
            console.error('Failed to initialize session:', error);
            this.showMessage('Failed to initialize session. Please refresh.', 'error');
        } finally {
            this.setLoading(false);
        }
    },
    
    async runQuery() {
        const query = document.getElementById('query').value.trim();
        
        if (!query) {
            this.showMessage('Please enter a SQL query', 'warning');
            return;
        }
        
        if (!this.sessionID) {
            await this.initializeSession();
            if (!this.sessionID) return;
        }
        
        try {
            this.setLoading(true);
            this.clearResults();
            
            const response = await fetch('/api/query', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ query })
            });
            
            const data = await response.json();
            
            if (data.error) {
                this.showMessage(`Error: ${data.error}`, 'error');
                return;
            }
            
            this.displayResults(data);
            
        } catch (error) {
            console.error('Query execution failed:', error);
            this.showMessage('Query failed to execute', 'error');
        } finally {
            this.setLoading(false);
        }
    },
    
    displayResults(data) {
        const resultDiv = document.getElementById('result');
        
        if (!data.rows || data.rows.length === 0) {
            resultDiv.innerHTML = '<p>Query executed successfully. No rows returned.</p>';
            return;
        }
        
        // Create table
        let html = `
            <div class="result-info">
                <p>${data.rows.length} row${data.rows.length !== 1 ? 's' : ''} returned</p>
            </div>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            ${data.columns.map(col => `<th>${this.escapeHtml(col)}</th>`).join('')}
                        </tr>
                    </thead>
                    <tbody>
        `;
        
        data.rows.forEach(row => {
            html += '<tr>';
            row.forEach(cell => {
                html += `<td>${this.escapeHtml(cell === null ? 'NULL' : cell.toString())}</td>`;
            });
            html += '</tr>';
        });
        
        html += `
                    </tbody>
                </table>
            </div>
        `;
        
        resultDiv.innerHTML = html;
    },
    
    async logout() {
        if (!this.sessionID) return;
        
        try {
            const response = await fetch('/api/logout', {
                method: 'POST'
            });
            
            if (response.ok) {
                this.sessionID = null;
                this.clearQuery();
                this.clearResults();
                this.showMessage('Session cleared', 'success');
                await this.initializeSession(); // Start fresh
            }
        } catch (error) {
            console.error('Logout failed:', error);
        }
    },
    
    setupEventListeners() {
        // Run query on button click
        document.getElementById('runQueryBtn').addEventListener('click', () => this.runQuery());
        
        // Run query on Ctrl+Enter in textarea
        document.getElementById('query').addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'Enter') {
                this.runQuery();
            }
        });
        
        // Clear session button
        document.getElementById('clearSessionBtn').addEventListener('click', () => this.logout());
        
        // Clear query button
        document.getElementById('clearQueryBtn').addEventListener('click', () => this.clearQuery());
    },
    
    setupUnloadHandler() {
        // Optional: Try to cleanup on page unload
        window.addEventListener('beforeunload', async (e) => {
            // Note: This is unreliable for async calls, but we try
            if (this.sessionID) {
                // Use synchronous version or just ignore errors
                navigator.sendBeacon('/api/logout');
            }
        });
    },
    
    clearQuery() {
        document.getElementById('query').value = '';
    },
    
    clearResults() {
        document.getElementById('result').innerHTML = '';
    },
    
    showMessage(message, type = 'info') {
        const messageDiv = document.getElementById('message');
        messageDiv.textContent = message;
        messageDiv.className = `message ${type}`;
        messageDiv.style.display = 'block';
        
        // Auto-hide after 5 seconds for success/info messages
        if (type === 'success' || type === 'info') {
            setTimeout(() => {
                messageDiv.style.display = 'none';
            }, 5000);
        }
    },
    
    setLoading(isLoading) {
        this.isLoading = isLoading;
        const button = document.getElementById('runQueryBtn');
        
        if (isLoading) {
            button.disabled = true;
            button.innerHTML = 'Running...';
        } else {
            button.disabled = false;
            button.innerHTML = 'Run Query';
        }
    },
    
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    QueryLabApp.init();
});