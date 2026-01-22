const QueryLabApp = {
    sessionID: null,
    isLoading: false,
    popupTimeout: null,
    progressInterval: null,

    async init() {
        console.log('QueryLab app initializing...');
        await this.initializeSession();
        this.setupEventListeners();
        this.setupUnloadHandler();
    },

    async initializeSession() {
        try {
            this.setLoading(true);

            const response = await fetch('/api/session', { method: 'POST', headers: { 'Content-Type': 'application/json' }});
            if (!response.ok) throw new Error(`HTTP ${response.status}`);

            const data = await response.json();
            this.sessionID = data.session_id;

            console.log('Session initialized:', this.sessionID);
            this.showPopup('Session ready', 'success');

        } catch (error) {
            console.error('Failed to initialize session:', error);
            this.showPopup('Failed to initialize session. Please refresh.', 'error');
        } finally {
            this.setLoading(false);
        }
    },

    async runQuery() {
        const query = document.getElementById('query').value.trim();
        if (!query) {
            this.showPopup('Please enter a SQL query', 'warning');
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
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query, session_id: this.sessionID })
            });

            if (!response.ok) throw new Error(`HTTP ${response.status}`);

            const data = await response.json();

            if (data.error) {
                this.showPopup(`Error: ${data.error}`, 'error');
                return;
            }

            this.displayResults(data);
            this.showPopup(`Query executed successfully`, 'success');

        } catch (error) {
            console.error('Query execution failed:', error);
            this.showPopup('Query failed to execute', 'error');
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

        let html = `
            <div class="result-info">
                <p>${data.rows.length} row${data.rows.length !== 1 ? 's' : ''} returned</p>
            </div>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>${data.columns.map(col => `<th>${this.escapeHtml(col)}</th>`).join('')}</tr>
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

        html += `</tbody></table></div>`;
        resultDiv.innerHTML = html;
    },

    async logout() {
        if (!this.sessionID) return;

        try {
            await fetch('/api/logout', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ session_id: this.sessionID })
            });

            this.sessionID = null;
            this.clearQuery();
            this.clearResults();
            this.showPopup('Session cleared', 'success');

            await this.initializeSession();
        } catch (error) {
            console.error('Logout failed:', error);
            this.showPopup('Failed to clear session', 'error');
        }
    },

    setupEventListeners() {
        document.getElementById('runQueryBtn').addEventListener('click', () => this.runQuery());
        document.getElementById('query').addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'Enter') this.runQuery();
        });
        document.getElementById('clearSessionBtn').addEventListener('click', () => this.logout());
        document.getElementById('clearQueryBtn').addEventListener('click', () => {
            this.clearQuery();
            this.showPopup('Query cleared', 'info');
        });
    },

    setupUnloadHandler() {
        window.addEventListener('beforeunload', () => {
            if (this.sessionID) {
                const payload = JSON.stringify({ session_id: this.sessionID });
                navigator.sendBeacon('/api/logout', payload);
            }
        });
    },

    clearQuery() {
        document.getElementById('query').value = '';
    },

    clearResults() {
        document.getElementById('result').innerHTML = '';
    },

    showPopup(message, type = 'info', duration = 5000) {
        const popup = document.getElementById('popup');
        popup.innerHTML = `<div class="popup-message">${message}</div><div class="popup-progress"></div>`;
        popup.className = `popup ${type}`;
        popup.style.display = 'block';

        const progressBar = popup.querySelector('.popup-progress');
        progressBar.style.width = '100%';

        const intervalTime = 50;
        let elapsed = 0;

        // Clear any previous intervals
        if (this.progressInterval) clearInterval(this.progressInterval);
        if (this.popupTimeout) clearTimeout(this.popupTimeout);

        this.progressInterval = setInterval(() => {
            elapsed += intervalTime;
            const percent = Math.max(0, 100 - (elapsed / duration) * 100);
            progressBar.style.width = percent + '%';
        }, intervalTime);

        this.popupTimeout = setTimeout(() => {
            popup.style.display = 'none';
            clearInterval(this.progressInterval);
        }, duration);
    },

    setLoading(isLoading) {
        this.isLoading = isLoading;
        const button = document.getElementById('runQueryBtn');

        if (isLoading) {
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Running...';
        } else {
            button.disabled = false;
            button.innerHTML = '<i class="fas fa-play"></i> Run Query (Ctrl+Enter)';
        }
    },

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

document.addEventListener('DOMContentLoaded', () => QueryLabApp.init());
