const debugContainer = document.getElementById('debug-container');
const refreshIntervalInput = document.getElementById('refresh-interval');
const setIntervalBtn = document.getElementById('set-interval-btn');
const expandAllBtn = document.getElementById('expand-all-btn');
const supervisorNameElement = document.getElementById('supervisor-name');
const searchInput = document.getElementById('search-input');
const searchPrevBtn = document.getElementById('search-prev-btn');
const searchNextBtn = document.getElementById('search-next-btn');
const searchResults = document.getElementById('search-results');

let refreshInterval = 1000; // Default to 1 second
let refreshIntervalId;
let previousData = null;
let isAllExpanded = true; // Start with all expanded
let searchMatches = [];
let currentSearchMatch = -1;
let lastSearchTerm = '';
let expandedState = {};
let currentSearchMatchPath = null;

function createTreeView(data, path = '') {
    const fragment = document.createDocumentFragment();

    for (const [key, value] of Object.entries(data)) {
        if (key === 'CollisionGrid') continue; // Skip CollisionGrid entirely

        const node = document.createElement('div');
        node.className = 'tree-node';
        const currentPath = path ? `${path}.${key}` : key;
        node.dataset.path = currentPath;

        const label = document.createElement('span');
        label.className = 'tree-label';
        
        if (typeof value === 'object' && value !== null) {
            const toggle = document.createElement('span');
            toggle.className = 'tree-toggle';
            toggle.textContent = expandedState[currentPath] !== false ? '▼' : '▶';
            label.appendChild(toggle);

            const keySpan = document.createElement('span');
            keySpan.className = 'tree-key';
            keySpan.textContent = key;
            label.appendChild(keySpan);

            const copyBtn = createCopyButton(currentPath);
            label.appendChild(copyBtn);

            node.appendChild(label);

            if (Array.isArray(value) && value.length > 100) {
                const arrayPreview = document.createElement('div');
                arrayPreview.className = 'large-dataset-notice';
                arrayPreview.textContent = `Array with ${value.length} items`;
                node.appendChild(arrayPreview);

                const childrenContainer = document.createElement('div');
                childrenContainer.style.display = expandedState[currentPath] !== false ? 'block' : 'none';
                childrenContainer.appendChild(createTreeView(Object.fromEntries(value.slice(0, 100).entries()), currentPath));
                node.appendChild(childrenContainer);
            } else {
                const childrenContainer = document.createElement('div');
                childrenContainer.style.display = expandedState[currentPath] !== false ? 'block' : 'none';
                childrenContainer.appendChild(createTreeView(value, currentPath));
                node.appendChild(childrenContainer);
            }

            label.addEventListener('click', (e) => {
                e.stopPropagation();
                toggleNode(currentPath);
            });
        } else {
            const keySpan = document.createElement('span');
            keySpan.className = 'tree-key';
            keySpan.textContent = `${key}: `;
            label.appendChild(keySpan);

            const valueSpan = document.createElement('span');
            valueSpan.className = `tree-value ${value === null ? 'null' : ''}`;
            valueSpan.textContent = JSON.stringify(value);
            label.appendChild(valueSpan);

            const copyBtn = createCopyButton(currentPath);
            label.appendChild(copyBtn);

            node.appendChild(label);
        }

        if (path === '' && fragment.children.length > 0) {
            const divider = document.createElement('div');
            divider.className = 'section-divider';
            fragment.appendChild(divider);
        }

        fragment.appendChild(node);
    }

    return fragment;
}

function toggleNode(path) {
    expandedState[path] = !expandedState[path];
    const node = debugContainer.querySelector(`[data-path="${path}"]`);
    if (node) {
        const label = node.querySelector('.tree-label');
        const toggle = label.querySelector('.tree-toggle');
        const childContainer = label.nextElementSibling;
        if (toggle && childContainer) {
            toggle.textContent = expandedState[path] ? '▼' : '▶';
            childContainer.style.display = expandedState[path] ? 'block' : 'none';
        }
    }
}

function createCopyButton(path) {
    const copyBtn = document.createElement('button');
    copyBtn.className = 'copy-btn';
    copyBtn.textContent = 'Copy';
    copyBtn.dataset.clipboardPath = path;
    copyBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        const textToCopy = JSON.stringify(getValueByPath(previousData, path), null, 2);
        navigator.clipboard.writeText(textToCopy).then(() => {
            copyBtn.textContent = 'Copied!';
            setTimeout(() => {
                copyBtn.textContent = 'Copy';
            }, 2000);
        });
    });
    return copyBtn;
}

function updateDebugContainer(data) {
    const newTree = createTreeView(data);
    debugContainer.innerHTML = '';
    debugContainer.appendChild(newTree);
    previousData = JSON.parse(JSON.stringify(data));
    updateExpandAllButton();
    if (lastSearchTerm) {
        performSearch(lastSearchTerm, false);
    }
}

function updateExpandAllButton() {
    expandAllBtn.querySelector('span').textContent = isAllExpanded ? 'Collapse All' : 'Expand All';
}

function fetchDebugData() {
    const urlParams = new URLSearchParams(window.location.search);
    const characterName = urlParams.get('characterName') || 'nullref';
    fetch(`/debug-data?characterName=${characterName}`)
        .then(response => response.json())
        .then(data => {
            delete data.CollisionGrid;
            updateDebugContainer(data);
            if (data.PlayerUnit && data.PlayerUnit.Name) {
                supervisorNameElement.textContent = `Supervisor: ${data.PlayerUnit.Name}`;
            } else {
                supervisorNameElement.textContent = `Supervisor: ${characterName}`;
            }
        })
        .catch(error => {
            console.error('Error:', error);
            debugContainer.innerHTML = '<p>Error fetching debug data</p>';
        });
}

function setRefreshInterval() {
    const newInterval = parseInt(refreshIntervalInput.value, 10) * 1000;
    if (newInterval && newInterval > 0) {
        refreshInterval = newInterval;
        clearInterval(refreshIntervalId);
        refreshIntervalId = setInterval(fetchDebugData, refreshInterval);
    }
}

function toggleExpandAll() {
    isAllExpanded = !isAllExpanded;
    const allNodes = debugContainer.querySelectorAll('.tree-node');
    allNodes.forEach(node => {
        const path = node.dataset.path;
        expandedState[path] = isAllExpanded;
        const label = node.querySelector('.tree-label');
        const toggle = label.querySelector('.tree-toggle');
        const childContainer = label.nextElementSibling;
        if (toggle && childContainer) {
            toggle.textContent = isAllExpanded ? '▼' : '▶';
            childContainer.style.display = isAllExpanded ? 'block' : 'none';
        }
    });
    updateExpandAllButton();
}

function getValueByPath(obj, path) {
    return path.split('.').reduce((acc, part) => acc && acc[part], obj);
}

function performSearch(searchTerm = searchInput.value, shouldScroll = true) {
    searchTerm = searchTerm.toLowerCase();
    lastSearchTerm = searchTerm;
    searchMatches = [];
    currentSearchMatch = -1;

    if (searchTerm) {
        searchRecursive(debugContainer, searchTerm);
        updateSearchResults();
        if (searchMatches.length > 0) {
            if (currentSearchMatchPath) {
                // Try to find the previous match in the new set of matches
                currentSearchMatch = searchMatches.findIndex(node => node.dataset.path === currentSearchMatchPath);
                if (currentSearchMatch === -1) {
                    currentSearchMatch = 0;
                }
            } else {
                currentSearchMatch = 0;
            }
            highlightAndScrollToCurrentSearchResult(shouldScroll);
        }
    } else {
        clearSearchHighlights();
        currentSearchMatchPath = null;
        updateSearchResults();
    }
}

function searchRecursive(node, searchTerm) {
    if (node.classList && node.classList.contains('tree-node')) {
        const label = node.querySelector('.tree-label');
        const text = label.textContent.toLowerCase();
        if (text.includes(searchTerm)) {
            searchMatches.push(node);
        }
    }
    for (const child of node.children) {
        searchRecursive(child, searchTerm);
    }
}

function updateSearchResults() {
    searchResults.textContent = searchMatches.length > 0 
        ? `${currentSearchMatch + 1}/${searchMatches.length}`
        : '0/0';
}

function goToNextSearchResult() {
    if (searchMatches.length > 0) {
        currentSearchMatch = (currentSearchMatch + 1) % searchMatches.length;
        highlightAndScrollToCurrentSearchResult(true);
    }
}

function goToPreviousSearchResult() {
    if (searchMatches.length > 0) {
        currentSearchMatch = (currentSearchMatch - 1 + searchMatches.length) % searchMatches.length;
        highlightAndScrollToCurrentSearchResult(true);
    }
}

function highlightAndScrollToCurrentSearchResult(shouldScroll = true) {
    clearSearchHighlights();
    if (currentSearchMatch >= 0 && currentSearchMatch < searchMatches.length) {
        const node = searchMatches[currentSearchMatch];
        if (node) {
            const label = node.querySelector('.tree-label');
            if (label) {
                label.classList.add('highlight');
                expandToNode(node);
                currentSearchMatchPath = node.dataset.path;
                if (shouldScroll) {
                    setTimeout(() => {
                        label.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    }, 100);
                }
            }
        }
    }
    updateSearchResults();
}

function clearSearchHighlights() {
    const highlightedElements = debugContainer.querySelectorAll('.highlight');
    highlightedElements.forEach(el => el.classList.remove('highlight'));
}

function expandToNode(node) {
    let current = node;
    while (current && current !== debugContainer) {
        if (current.classList && current.classList.contains('tree-node')) {
            const path = current.dataset.path;
            if (path) {
                expandedState[path] = true;
                const label = current.querySelector('.tree-label');
                const toggle = label && label.querySelector('.tree-toggle');
                const childContainer = label && label.nextElementSibling;
                if (toggle && childContainer) {
                    toggle.textContent = '▼';
                    childContainer.style.display = 'block';
                }
            }
        }
        current = current.parentElement;
    }
}

function createCopyDataButton() {
    const copyDataBtn = document.getElementById('copy-data-btn');
    copyDataBtn.addEventListener('click', () => {
        const textToCopy = JSON.stringify(previousData, null, 2);
        navigator.clipboard.writeText(textToCopy).then(() => {
            const originalText = copyDataBtn.innerHTML;
            copyDataBtn.innerHTML = 'Copied!';
            setTimeout(() => {
                copyDataBtn.innerHTML = originalText;
            }, 2000);
        });
    });
}

// Event Listeners
setIntervalBtn.addEventListener('click', setRefreshInterval);
expandAllBtn.addEventListener('click', toggleExpandAll);
searchInput.addEventListener('input', () => performSearch());
searchNextBtn.addEventListener('click', goToNextSearchResult);
searchPrevBtn.addEventListener('click', goToPreviousSearchResult);

// Initialize
createCopyDataButton();
fetchDebugData();
refreshIntervalId = setInterval(fetchDebugData, refreshInterval);