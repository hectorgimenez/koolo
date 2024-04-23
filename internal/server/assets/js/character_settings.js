window.onload = function () {
    let enabled_runs_ul = document.getElementById('enabled_runs')
    let disabled_runs_ul = document.getElementById('disabled_runs')

    new Sortable(enabled_runs_ul, {
        group: 'runs',
        animation: 150
    });

    new Sortable(disabled_runs_ul, {
        group: 'runs',
        animation: 150
    });
}