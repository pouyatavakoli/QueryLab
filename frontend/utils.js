const QueryLabUtils = {
    // Format SQL query for display
    formatSql(sql) {
        // Simple SQL formatting - in a real app, use a proper SQL formatter library
        return sql
            .replace(/\b(SELECT|FROM|WHERE|JOIN|LEFT|RIGHT|INNER|OUTER|GROUP BY|ORDER BY|LIMIT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP)\b/gi, '\n$1')
            .replace(/,(\s*\w)/g, ',\n$1')
            .trim();
    },
    
    // Copy text to clipboard
    copyToClipboard(text) {
        return navigator.clipboard.writeText(text)
            .then(() => true)
            .catch(err => {
                console.error('Failed to copy:', err);
                return false;
            });
    },
    
    // Download results as CSV
    downloadCsv(columns, rows, filename = 'querylab_results.csv') {
        const csvContent = [
            columns.join(','),
            ...rows.map(row => row.map(cell => {
                if (cell === null) return '';
                const str = cell.toString();
                // Escape quotes and wrap in quotes if contains comma or quote
                if (str.includes(',') || str.includes('"') || str.includes('\n')) {
                    return `"${str.replace(/"/g, '""')}"`;
                }
                return str;
            }).join(','))
        ].join('\n');
        
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const link = document.createElement('a');
        const url = URL.createObjectURL(blob);
        
        link.setAttribute('href', url);
        link.setAttribute('download', filename);
        link.style.visibility = 'hidden';
        
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    },
    
    // Get query from URL parameter (useful for sharing queries)
    getQueryFromUrl() {
        const params = new URLSearchParams(window.location.search);
        const query = params.get('q');
        return query ? decodeURIComponent(query) : null;
    },
    
    // Set query in URL parameter
    setQueryInUrl(query) {
        const url = new URL(window.location);
        if (query) {
            url.searchParams.set('q', encodeURIComponent(query));
        } else {
            url.searchParams.delete('q');
        }
        window.history.pushState({}, '', url);
    }
};

function showPopup(message, type = "info", duration = 4000) {
    const popup = document.getElementById("popup");
    popup.textContent = message;
    popup.className = `popup ${type}`;
    popup.style.display = "block";

    setTimeout(() => {
        popup.style.display = "none";
    }, duration);
}
