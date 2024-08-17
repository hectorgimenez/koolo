window.onload = function () {
    let enabled_runs_ul = document.getElementById('enabled_runs')
    let disabled_runs_ul = document.getElementById('disabled_runs')
    let searchInput = document.getElementById('search-disabled-runs');

    new Sortable(enabled_runs_ul, {
        group: 'runs',
        animation: 150,
        onSort: function (evt) {
            updateEnabledRunsHiddenField();
        }
    });

    new Sortable(disabled_runs_ul, {
        group: 'runs',
        animation: 150
    });

    searchInput.addEventListener('input', function() {
        filterDisabledRuns(searchInput.value);
    });
    
    clearButton.addEventListener('click', function() {
        confirmClearEnabledRuns();
    });

    updateEnabledRunsHiddenField();
}

function updateEnabledRunsHiddenField() {
    let listItems = document.querySelectorAll('#enabled_runs li');
    let values = Array.from(listItems).map(function(item) {
        return item.getAttribute("value");
    });
    document.getElementById('gameRuns').value = JSON.stringify(values);
}

function filterDisabledRuns(searchTerm) {
    let listItems = document.querySelectorAll('#disabled_runs li');
    searchTerm = searchTerm.toLowerCase();
    listItems.forEach(function(item) {
        let runName = item.getAttribute("value").toLowerCase();
        if (runName.includes(searchTerm)) {
            item.style.display = '';
        } else {
            item.style.display = 'none';
        }
    });
}

function confirmClearEnabledRuns() {
    if (confirm("Are you sure you want to clear all enabled runs?")) {
        clearEnabledRuns();
    }
}

function clearEnabledRuns() {
    let enabledRunsUl = document.getElementById('enabled_runs');
    enabledRunsUl.innerHTML = '';
    updateEnabledRunsHiddenField();
}