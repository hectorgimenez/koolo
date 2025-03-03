window.onload = function () {
    let enabled_runs_ul = document.getElementById('enabled_runs');
    let disabled_runs_ul = document.getElementById('disabled_runs');
    let searchInput = document.getElementById('search-disabled-runs');

    new Sortable(enabled_runs_ul, {
        group: 'runs',
        animation: 150,
        onSort: function (evt) {
            updateEnabledRunsHiddenField();
        },
        onAdd: function (evt) {
            updateButtonForEnabledRun(evt.item);
        }
    });

    new Sortable(disabled_runs_ul, {
        group: 'runs',
        animation: 150,
        onAdd: function (evt) {
            updateButtonForDisabledRun(evt.item);
        }
    });

    searchInput.addEventListener('input', function () {
        filterDisabledRuns(searchInput.value);
    });

    // Add event listeners for add and remove buttons
    document.addEventListener('click', function (e) {
        if (e.target.closest('.remove-run')) {
            e.preventDefault();
            const runElement = e.target.closest('li');
            moveRunToDisabled(runElement);
        } else if (e.target.closest('.add-run')) {
            e.preventDefault();
            const runElement = e.target.closest('li');
            moveRunToEnabled(runElement);
        }
    });

    updateEnabledRunsHiddenField();
}

function updateEnabledRunsHiddenField() {
    let listItems = document.querySelectorAll('#enabled_runs li');
    let values = Array.from(listItems).map(function (item) {
        return item.getAttribute("value");
    });
    document.getElementById('gameRuns').value = JSON.stringify(values);
}

function filterDisabledRuns(searchTerm) {
    let listItems = document.querySelectorAll('#disabled_runs li');
    searchTerm = searchTerm.toLowerCase();
    listItems.forEach(function (item) {
        let runName = item.getAttribute("value").toLowerCase();
        if (runName.includes(searchTerm)) {
            item.style.display = '';
        } else {
            item.style.display = 'none';
        }
    });
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

function moveRunToDisabled(runElement) {
    const disabledRunsUl = document.getElementById('disabled_runs');
    updateButtonForDisabledRun(runElement);
    disabledRunsUl.appendChild(runElement);
    updateEnabledRunsHiddenField();
}

function moveRunToEnabled(runElement) {
    const enabledRunsUl = document.getElementById('enabled_runs');
    updateButtonForEnabledRun(runElement);
    enabledRunsUl.appendChild(runElement);
    updateEnabledRunsHiddenField();
}

function updateButtonForEnabledRun(runElement) {
    const button = runElement.querySelector('button');
    button.classList.remove('add-run');
    button.classList.add('remove-run');
    button.title = "Remove run";
    button.innerHTML = '<i class="bi bi-dash"></i>';
}

function updateButtonForDisabledRun(runElement) {
    const button = runElement.querySelector('button');
    button.classList.remove('remove-run');
    button.classList.add('add-run');
    button.title = "Add run";
    button.innerHTML = '<i class="bi bi-plus"></i>';
}

document.addEventListener('DOMContentLoaded', function () {
    const schedulerEnabled = document.querySelector('input[name="schedulerEnabled"]');
    const schedulerSettings = document.getElementById('scheduler-settings');
    const characterClassSelect = document.querySelector('select[name="characterClass"]');
    const berserkerBarbOptions = document.querySelector('.berserker-barb-options');
    const novaSorceressOptions = document.querySelector('.nova-sorceress-options');
    const bossStaticThresholdInput = document.getElementById('novaBossStaticThreshold');
    const mosaicAssassinOptions = document.querySelector('.mosaic-assassin-options');
    const useTeleportCheckbox = document.getElementById('characterUseTeleport');
    const clearPathDistContainer = document.getElementById('clearPathDistContainer');
    const clearPathDistInput = document.getElementById('clearPathDist');
    const clearPathDistValue = document.getElementById('clearPathDistValue');

    if (bossStaticThresholdInput) {
        bossStaticThresholdInput.addEventListener('input', handleBossStaticThresholdChange);
    }

    function toggleSchedulerVisibility() {
        schedulerSettings.style.display = schedulerEnabled.checked ? 'grid' : 'none';
    }

    function updateCharacterOptions() {
        const selectedClass = characterClassSelect.value;
        const noSettingsMessage = document.getElementById('no-settings-message');
        const berserkerBarbOptions = document.querySelector('.berserker-barb-options');
        const novaSorceressOptions = document.querySelector('.nova-sorceress-options');
        const mosaicAssassinOptions = document.querySelector('.mosaic-assassin-options');
        // Hide all options first
        berserkerBarbOptions.style.display = 'none';
        novaSorceressOptions.style.display = 'none';
        mosaicAssassinOptions.style.display = 'none';
        noSettingsMessage.style.display = 'none';
        
        // Show relevant options based on class
        if (selectedClass === 'berserker') {
            berserkerBarbOptions.style.display = 'block';
        } else if (selectedClass === 'nova' || selectedClass === 'lightsorc') {
            novaSorceressOptions.style.display = 'block';
            updateNovaSorceressOptions();
        } else if (selectedClass === 'mosaic') {
            mosaicAssassinOptions.style.display = 'block';
        } else {
            noSettingsMessage.style.display = 'block';
        }
    }
    function toggleClearPathVisibility() {
        if (useTeleportCheckbox && clearPathDistContainer) {
            if (useTeleportCheckbox.checked) {
                clearPathDistContainer.style.display = 'none';
            } else {
                clearPathDistContainer.style.display = 'block';
            }
        }
    }

    // Update the displayed value when the slider changes
    function updateClearPathValue() {
        if (clearPathDistInput && clearPathDistValue) {
            clearPathDistValue.textContent = clearPathDistInput.value;
        }
    }

    // Set up event listeners
    if (useTeleportCheckbox) {
        useTeleportCheckbox.addEventListener('change', toggleClearPathVisibility);
        // Initialize visibility
        toggleClearPathVisibility();
    }

    if (clearPathDistInput) {
        clearPathDistInput.addEventListener('input', updateClearPathValue);
        // Initialize value display
        updateClearPathValue();
    }
    
    function updateNovaSorceressOptions() {
        const selectedDifficulty = document.getElementById('gameDifficulty').value;
        updateBossStaticThresholdMin(selectedDifficulty);
        handleBossStaticThresholdChange();
    }
    
    function updateBossStaticThresholdMin(difficulty) {
        const input = document.getElementById('novaBossStaticThreshold');
        let minValue;
        switch(difficulty) {
            case 'normal':
                minValue = 1;
                break;
            case 'nightmare':
                minValue = 33;
                break;
            case 'hell':
                minValue = 50;
                break;
            default:
                minValue = 65;
        }
        input.min = minValue;

        // Ensure the current value is not less than the new minimum
        if (parseInt(input.value) < minValue) {
            input.value = minValue;
        }
    }

    characterClassSelect.addEventListener('change', updateCharacterOptions);
    document.getElementById('gameDifficulty').addEventListener('change', function() {
        if (characterClassSelect.value === 'nova' || characterClassSelect.value === 'lightsorc') {
            updateNovaSorceressOptions();
        }
    });

    characterClassSelect.addEventListener('change', updateCharacterOptions);
    updateCharacterOptions(); // Call this initially to set the correct state

    // Set initial state
    toggleSchedulerVisibility();
    updateNovaSorceressOptions();

    schedulerEnabled.addEventListener('change', toggleSchedulerVisibility);

    document.querySelectorAll('.add-time-range').forEach(button => {
        button.addEventListener('click', function () {
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

    document.addEventListener('click', function (e) {
        if (e.target.closest('.remove-time-range')) {
            e.target.closest('.time-range').remove();
        }
    });

    document.getElementById('tzTrackAll').addEventListener('change', function (e) {
        document.querySelectorAll('.tzTrackCheckbox').forEach(checkbox => {
            checkbox.checked = e.target.checked;
        });
    });
});

function handleBossStaticThresholdChange() {
    const input = document.getElementById('novaBossStaticThreshold');
    const selectedDifficulty = document.getElementById('gameDifficulty').value;
    let minValue;
    switch(selectedDifficulty) {
        case 'normal':
            minValue = 1;
            break;
        case 'nightmare':
            minValue = 33;
            break;
        case 'hell':
            minValue = 50;
            break;
        default:
            minValue = 65;
    }

    let value = parseInt(input.value);
    if (isNaN(value) || value < minValue) {
        value = minValue;
    } else if (value > 100) {
        value = 100;
    }
    input.value = value;
}