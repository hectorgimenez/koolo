window.onload = function () {
    let enabled_runs_ul = document.getElementById('enabled_runs')
    let disabled_runs_ul = document.getElementById('disabled_runs')

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

    updateEnabledRunsHiddenField();
}

function updateEnabledRunsHiddenField() {
    let listItems = document.querySelectorAll('#enabled_runs li');
    let values = Array.from(listItems).map(function(item) {
        return item.getAttribute("value");
    });
    document.getElementById('gameRuns').value = JSON.stringify(values);
}