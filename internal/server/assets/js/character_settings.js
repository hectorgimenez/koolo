window.onload = function () {
    let enabled_runs_ul = document.getElementById('enabled_runs')
    let disabled_runs_ul = document.getElementById('disabled_runs')
    let searchInput = document.getElementById('search-disabled-runs');
    let clearButton = document.getElementById('clear-enabled-runs');
    
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

function checkLevelingProfile() {
    const levelingProfiles = [
        "sorceress_leveling_hydraorb",
        "sorceress_leveling_lightning",
        "sorceress_leveling",
        "paladin_leveling"
    ];
    const characterClass = document.getElementById('characterClass').value;

    if (levelingProfiles.includes(characterClass)) {
        const confirmation = confirm("This profile requires the leveling run profile, would you like to clear enabled run profiles and select the leveling profile?");
        if (confirmation) {
            clearEnabledRuns();
            selectLevelingProfile();
        }
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const schedulerEnabled = document.querySelector('input[name="schedulerEnabled"]');
    const schedulerSettings = document.getElementById('scheduler-settings');

    function toggleSchedulerVisibility() {
        schedulerSettings.style.display = schedulerEnabled.checked ? 'grid' : 'none';
    }

    // Set initial state
    toggleSchedulerVisibility();

    schedulerEnabled.addEventListener('change', toggleSchedulerVisibility);

    document.querySelectorAll('.add-time-range').forEach(button => {
        button.addEventListener('click', function() {
            const day = this.dataset.day;
            const timeRangesDiv = this.previousElementSibling;
            if (timeRangesDiv) {
                const newTimeRange = document.createElement('div');
                newTimeRange.className = 'time-range';
                newTimeRange.innerHTML = `
                    <input type="time" name="scheduler[${day}][start][]" required>
                    <span>to</span>
                    <input type="time" name="scheduler[${day}][end][]" required>
                    <button type="button" class="remove-time-range"><i class="bi bi-trash"></i></button>
                `;
                timeRangesDiv.appendChild(newTimeRange);
            }
        });
    });

    document.addEventListener('click', function(e) {
        if (e.target.closest('.remove-time-range')) {
            e.target.closest('.time-range').remove();
        }
    });
});