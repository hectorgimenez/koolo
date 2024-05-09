document.addEventListener("DOMContentLoaded", function() {
    const container = document.getElementById('jsoneditor');
    const options = {
        mode: 'tree'
    };

    const editor = new JSONEditor(container, options);

    function fetchData() {
        const characterName = new URLSearchParams(window.location.search).get('characterName');
        if (!characterName) {
            console.error('Character name is missing from URL');
            return;
        }
    
        const endpoint = `/debug-data?characterName=${encodeURIComponent(characterName)}`;
    
        fetch(endpoint)
            .then(response => response.json())
            .then(data => {
                editor.update(data);
            })
            .catch(error => console.error('Error loading data:', error));
    }

    fetchData();
    setInterval(fetchData, 1000); // Adjusted for less frequent updates
});

function fetchData() {
    const characterName = new URLSearchParams(window.location.search).get('characterName');
    if (!characterName) {
        console.error('Character name is missing from URL');
        return;
    }

    const endpoint = `/debug-data?characterName=${encodeURIComponent(characterName)}`;

    fetch(endpoint)
        .then(response => response.json())
        .then(data => {
            editor.update(data);
        })
        .catch(error => console.error('Error loading data:', error));
}
