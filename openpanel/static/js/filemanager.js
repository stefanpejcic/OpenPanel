var selectedRows = [];
var sveselect = 0;
var wasSelecting = false;
var isSelecting = false;
var startRowIndex = null;
var endRowIndex = null;

// HELPERS
function setClassAll(selector, className) {
    document.querySelectorAll(selector).forEach(el => el.className = className);
}

function setDisabledAll(selector, disabled) {
    document.querySelectorAll(selector).forEach(el => el.disabled = disabled);
}

function setButtonsState(selectors, enabled, customClass = FM.btn.base) {
    document.querySelectorAll(selectors).forEach(el => {
        el.disabled = !enabled;
        el.className = customClass;
    });
}

function resetSelectAllBtn() {
    document.getElementById('SelectAll-button').className = FM.btn.baseFit;
}

function setSelectAllLabel(deselect) {
    const icon = deselect ? 'bi-hand-index-fill' : 'bi-hand-index';
    const text = deselect ? FM.i18n.deselect : FM.i18n.selectAll;
    document.getElementById('spanAll').innerHTML = `<i class="bi ${icon}"></i> ${text}`;
    resetSelectAllBtn();
}

// ENABLE/DISABLE BUTTONS

function disableButtons() {
    setDisabledAll('#viewButton, #editButton, #renameButton, #copyButton, #moveButton, #downloadButton, #deleteButton, #permButton, #compressButton, #extractButton', true);
    setClassAll('#viewButton, #editButton, #renameButton, #copyButton, #moveButton, #permButton, #compressButton, #extractButton', FM.btn.base);
    setClassAll('#downloadButton', FM.btn.download);
    setClassAll('#deleteButton', FM.btn.red);
}

function enableDisableButtons() {
    const table = document.getElementById('filemanager_table');
    const row = table.querySelector('tbody tr.selected-row');
    const fileName = row ? row.dataset.file : undefined;
    const fileType = row ? row.dataset.type : undefined;
    const fileSize = row ? (parseInt(row.getAttribute('data-size'), 10) || 0) : 0;

    const isLargeForEdit     = fileSize > FM.editSizeLimit;
    const isLargeForView     = fileSize > FM.viewSizeLimit;
    const isLargeForDownload = fileSize > FM.downloadSizeLimit;

    setButtonsState('#downloadButton', false, FM.btn.download);
    setButtonsState('#deleteButton',   false, FM.btn.red);
    setButtonsState('#renameButton, #permButton, #viewButton, #editButton, #extractButton', false);

    const count = selectedRows.length;

    if (count === 0) {
        setButtonsState('#copyButton, #moveButton, #compressButton', false);
        return;
    }

    if (count > 1) {
        setButtonsState('#compressButton, #copyButton, #moveButton, #permButton', true);
        setButtonsState('#deleteButton', true, FM.btn.red);
        setButtonsState('#renameButton, #viewButton, #editButton, #extractButton', false);
        setButtonsState('#downloadButton', false, FM.btn.download);
        return;
    }

    // single selection
    if (fileType === 'directory') {
        setButtonsState('#renameButton, #copyButton, #moveButton, #permButton, #compressButton', true);
        setButtonsState('#deleteButton', true, FM.btn.red);
        setButtonsState('#viewButton, #editButton, #extractButton', false);
        setButtonsState('#downloadButton', false, FM.btn.download);
        return;
    }

    // single file
    const hasExtension = ext => {
        if (!fileName) return false;
        if (!ext.startsWith('.')) return fileName.toLowerCase().includes(ext.toLowerCase());
        const dot = fileName.lastIndexOf('.');
        if (dot === -1) return false;
        return fileName.slice(dot).toLowerCase() === ext.toLowerCase();
    };
    const isMatch    = list => list.some(hasExtension);
    const isEditable = isMatch(FM.extensions);
    const isImage    = isMatch(FM.images);
    const isArchive  = isMatch(FM.archives);

    setButtonsState('#renameButton, #copyButton, #moveButton, #permButton, #compressButton', true);
    setButtonsState('#deleteButton', true, FM.btn.red);

    if (!isLargeForDownload) {
        setButtonsState('#downloadButton', true, FM.btn.download);
    }

    if (isArchive) {
        setButtonsState('#extractButton', true);
        setButtonsState('#viewButton, #editButton', false);
    } else if (isImage && !isLargeForEdit) {
        document.getElementById('viewButton').setAttribute('href',
            `/file-manager/view-file?filename=${encodeURIComponent(fileName)}&path_param=${encodeURIComponent(FM.pathParam)}`
        );
        setButtonsState('#viewButton', true);
        setButtonsState('#editButton, #extractButton', false);
    } else if (isEditable) {
        setButtonsState('#editButton', !isLargeForEdit);
        setButtonsState('#viewButton', !isLargeForView);
        setButtonsState('#extractButton', false);
    } else {
        setButtonsState('#viewButton, #editButton, #extractButton', false);
    }
}

// DISPLAY SELECTED

function fadeIn(el, duration) {
    el.classList.remove('hidden');
    el.style.transition = `opacity ${duration}ms`;
    el.style.opacity = 0;
    requestAnimationFrame(() => el.style.opacity = 1);
}

function fadeOut(el, duration, callback) {
    el.style.transition = `opacity ${duration}ms`;
    el.style.opacity = 0;
    setTimeout(() => { if (callback) callback(); }, duration);
}

function updateSelectedOptionsDisplay() {
    const selectedOptions = document.getElementById('selectedOptions');
    const selectedRowsDisplay = document.getElementById('selectedRows');
    if (!selectedOptions || !selectedRowsDisplay) return;

    if (selectedRows.length > 0) {
        selectedRowsDisplay.textContent = `${selectedRows.length} ${FM.i18n.selected}`;
        fadeIn(selectedOptions, 150);
    } else {
        fadeOut(selectedOptions, 150, () => selectedOptions.classList.add('hidden'));
    }
}

// SORT TABLE

const sortState = { columnIndex: -1, dir: 'none' };

function parseLsDate(str) {
    if (!str) return 0;
    str = str.trim();
    const currentYear = new Date().getFullYear();

    const withTime = str.match(/^(\w{3})\s+(\d+)\s+(\d+):(\d+)$/);
    if (withTime) {
        const [, mon, day, hour, min] = withTime;
        return new Date(`${mon} ${day} ${currentYear} ${hour}:${min}`).getTime();
    }

    const withYear = str.match(/^(\w{3})\s+(\d+)\s+(\d{4})$/);
    if (withYear) {
        const [, mon, day, year] = withYear;
        return new Date(`${mon} ${day} ${year}`).getTime();
    }

    return 0;
}

function parseSize(str) {
    if (!str) return 0;
    str = str.trim();
    if (/^\d+$/.test(str)) return parseInt(str, 10);
    const match = str.match(/^([\d.]+)\s*([BKMGT])/i);
    if (!match) return 0;
    const map = { B: 1, K: 1024, M: 1024**2, G: 1024**3, T: 1024**4 };
    return parseFloat(match[1]) * (map[match[2].toUpperCase()] || 1);
}

function sortTable(tableId, columnIndex) {
    const table = document.getElementById(tableId);
    const tbody = table.querySelector('tbody');
    const rows  = Array.from(tbody.querySelectorAll('tr.clickable-row'));
    const allThs = Array.from(table.querySelectorAll('thead tr th'));
    const th = allThs[columnIndex];
    if (!th) return;

    const currentDir = (sortState.columnIndex === columnIndex) ? sortState.dir : 'none';
    const newDir = currentDir === 'desc' ? 'asc' : 'desc'; // desc first

    sortState.columnIndex = columnIndex;
    sortState.dir = newDir;

    allThs.forEach(h => h.querySelector('.sort-icon')?.remove());
    const icon = document.createElement('span');
    icon.className = 'sort-icon ml-1 text-xs text-gray-400';
    icon.textContent = newDir === 'asc' ? '▲' : '▼';
    (th.querySelector('span') || th).appendChild(icon);

    rows.sort((a, b) => {
        const aCell = a.querySelectorAll('td')[columnIndex];
        const bCell = b.querySelectorAll('td')[columnIndex];

        if (a.dataset.type === 'directory' && b.dataset.type !== 'directory') return -1;
        if (a.dataset.type !== 'directory' && b.dataset.type === 'directory') return 1;

        let aVal, bVal;

        if (columnIndex === 1) {
            aVal = a.dataset.type === 'file' ? parseInt(a.dataset.size || '0', 10) : parseSize(aCell?.textContent.trim());
            bVal = b.dataset.type === 'file' ? parseInt(b.dataset.size || '0', 10) : parseSize(bCell?.textContent.trim());
            return newDir === 'asc' ? aVal - bVal : bVal - aVal;
        }

        if (columnIndex === 2) {
            aVal = parseLsDate(a.dataset.date);
            bVal = parseLsDate(b.dataset.date);
            return newDir === 'asc' ? aVal - bVal : bVal - aVal;
        }

        if (aCell?.classList.contains('permissions-cell')) {
            aVal = a.dataset.permissions || '';
            bVal = b.dataset.permissions || '';
            return newDir === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
        }

        aVal = aCell?.textContent.trim().toLowerCase() || '';
        bVal = bCell?.textContent.trim().toLowerCase() || '';
        return newDir === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
    });

    const goUpRow = tbody.querySelector('tr:not(.clickable-row)');
    rows.forEach(r => tbody.appendChild(r));
    if (goUpRow) tbody.insertBefore(goUpRow, tbody.firstChild);
}

// ROW SELECT

function toggleClasses(el, classString) {
    classString.split(/\s+/).filter(Boolean).forEach(c => el.classList.toggle(c));
}

function initTableSelection() {
    const table = document.getElementById('filemanager_table');

    // click row
    table.querySelectorAll('tbody tr.clickable-row').forEach(row => {
        row.addEventListener('click', function(event) {
            const rowIndex = this.dataset.rowIndex;

            if (event.ctrlKey || event.metaKey) {
                toggleClasses(this, 'selected-row bg-gray-200 dark:bg-gray-900');
                const idx = selectedRows.indexOf(rowIndex);
                if (idx === -1) selectedRows.push(rowIndex);
                else selectedRows.splice(idx, 1);
                sveselect = 0;
                setSelectAllLabel(false);
            } else {
                table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                    r.classList.remove('selected-row', 'bg-gray-200', 'dark:bg-gray-900')
                );
                toggleClasses(this, 'selected-row bg-gray-200 dark:bg-gray-900');
                selectedRows = [rowIndex];
            }
            enableDisableButtons();
            updateSelectedOptionsDisplay();
        });
    });

    // sortable headers
    document.querySelectorAll('#filemanager_table th.sortable').forEach(th => {
        th.addEventListener('click', function(event) {
            event.stopPropagation();
            const allThs = Array.from(this.closest('thead').querySelectorAll('th'));
            const idx = allThs.indexOf(this);
            sortTable('filemanager_table', idx);
        });
    });

    // drag select
    table.addEventListener('mousedown', function(event) {
        if (event.ctrlKey || event.metaKey || event.button !== 0) return;

        isSelecting = true;
        wasSelecting = true;
        sveselect = 0;
        setSelectAllLabel(false);
        startRowIndex = endRowIndex = null;

        const startX = event.pageX, startY = event.pageY;
        const rect = document.createElement('div');
        rect.classList.add('selection-rectangle');
        rect.style.cssText = `top:${startY}px;left:${startX}px;user-select:none`;
        document.body.appendChild(rect);

        const onMouseMove = event => {
            if (!isSelecting) return;
            const endX = event.pageX, endY = event.pageY;
            rect.style.width  = Math.abs(endX - startX) + 'px';
            rect.style.height = Math.abs(endY - startY) + 'px';
            rect.style.left   = Math.min(startX, endX) + 'px';
            rect.style.top    = Math.min(startY, endY) + 'px';

            table.querySelectorAll('tbody tr.clickable-row').forEach(row => {
                const r = row.getBoundingClientRect();
                const rLeft = r.left + window.scrollX, rTop = r.top + window.scrollY;
                const intersects = !(
                    rLeft + row.offsetWidth  < Math.min(startX, endX) ||
                    rLeft                    > Math.max(startX, endX) ||
                    rTop  + row.offsetHeight < Math.min(startY, endY) ||
                    rTop                     > Math.max(startY, endY)
                );
                const rowIndex = row.dataset.rowIndex;
                if (intersects) {
                    row.classList.add('selected-row', 'bg-gray-200', 'dark:bg-gray-900');
                    if (!selectedRows.includes(rowIndex)) selectedRows.push(rowIndex);
                } else {
                    row.classList.remove('selected-row', 'bg-gray-200', 'dark:bg-gray-900');
                    const i = selectedRows.indexOf(rowIndex);
                    if (i !== -1) selectedRows.splice(i, 1);
                }
            });
            enableDisableButtons();
            updateSelectedOptionsDisplay();
        };

        const onMouseUp = () => {
            isSelecting = false;
            wasSelecting = true;
            rect.remove();
            document.removeEventListener('mousemove', onMouseMove);
            document.removeEventListener('mouseup', onMouseUp);
            enableDisableButtons();
            updateSelectedOptionsDisplay();
        };

        document.addEventListener('mousemove', onMouseMove);
        document.addEventListener('mouseup', onMouseUp);
    });

    // click outside to deselect
    document.addEventListener('click', event => {
        const t = event.target;
        if (
            !t.closest('#filemanager_table') &&
            !t.closest('#mainButtons') &&
            !t.closest('#selectedOptions') &&
            !t.closest('.modal.fade') &&
            !t.closest('.btn-group') &&
            !t.closest('.fmdrawer') &&
            !wasSelecting
        ) {
            table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                r.classList.remove('selected-row', 'bg-gray-200', 'dark:bg-gray-900')
            );
            selectedRows = [];
            sveselect = 0;
            setSelectAllLabel(false);
            disableButtons();
            updateSelectedOptionsDisplay();
        }
        wasSelecting = false;
    });
}

// SELECT ALL BUTTON

function initSelectAll() {
    document.getElementById('SelectAll-button').addEventListener('click', () => {
        const table = document.getElementById('filemanager_table');
        const rows  = table.querySelectorAll('tbody tr.clickable-row');

        if (!sveselect) {
            rows.forEach(r => r.classList.add('selected-row', 'bg-gray-200', 'dark:bg-gray-900'));
            selectedRows = Array.from(rows).map((_, i) => String(i));
            sveselect = 1;
            setSelectAllLabel(true);
        } else {
            rows.forEach(r => r.classList.remove('selected-row', 'bg-gray-200', 'dark:bg-gray-900'));
            selectedRows = [];
            sveselect = 0;
            setSelectAllLabel(false);
            disableButtons();
        }
        enableDisableButtons();
        updateSelectedOptionsDisplay();
    });
}

// CALCULATE SIZE

function initDirectorySizeCalculate() {
    document.querySelectorAll('.get-folder-size-button').forEach(button => {
        button.addEventListener('click', function(event) {
            event.preventDefault();
            if (this.dataset.loading === 'true') return;

            const cell = this.closest('td');
            if (!cell) return;

            this.dataset.loading = 'true';
            this.style.pointerEvents = 'none';
            this.textContent = FM.i18n.calculating;

            const folderName = this.getAttribute('data-folder');
            const currentUrl = window.location.href.split(/[?#]/)[0];
            const pathParam  = currentUrl.split('/files/')[1];
            const url = pathParam
                ? `/json/directory-size?folder=${pathParam}/${encodeURIComponent(folderName)}`
                : `/json/directory-size?folder=${encodeURIComponent(folderName)}`;

            fetch(url, { headers: { 'X-CSRF-Token': csrf_token } })
                .then(r => r.json())
                .then(data => { cell.textContent = data.size; })
                .catch(error => {
                    showToast(`Error fetching folder size: ${error.message || error}`, 'error');
                    if (button.isConnected) {
                        button.dataset.loading = 'false';
                        button.style.pointerEvents = '';
                        button.textContent = FM.i18n.calculate;
                    }
                });
        });
    });
}

// ACTION BUTTONS

function initActionButtons() {
    // COPY
    document.getElementById('copyButton').addEventListener('click', () => {
        const items = getSelectedItems();
        populateDrawerList('#copyDrawer .selected-items-list', items);

        const confirmBtn = document.getElementById('copyConfirmButton');
        const newBtn = confirmBtn.cloneNode(true);
        confirmBtn.parentNode.replaceChild(newBtn, confirmBtn);

        newBtn.addEventListener('click', () => {
            const dest = document.getElementById('copyDestination').value;
            const total = items.length;
            let done = 0;
            showProgressBar('copy');

            items.forEach(({ fileName, itemType }) => {
                fetch(`/file-manager/copy?item_name=${encodeURIComponent(fileName)}&path_param=${encodeURIComponent(FM.pathParam)}&item_type=${encodeURIComponent(itemType)}&destination_path=${encodeURIComponent(dest)}`, {
                    method: 'POST', headers: { 'X-CSRF-Token': csrf_token }
                })
                .then(r => r.json())
                .then(data => {
                    if (data.success) {
                        done++;
                        updateProgressBar('copy', done, total);
                        if (done === total) finishProgressBar('copy', 'Copy complete!', 'Copy completed successfully! Reloading page.');
                    } else {
                        showToast(data.error || 'Unknown error', 'error');
                    }
                })
                .catch(e => showToast(e.message, 'error'));
            });
        });
    });

    // MOVE
    document.getElementById('moveButton').addEventListener('click', () => {
        const items = getSelectedItems();
        populateDrawerList('#moveDrawer .selected-items-list', items);

        document.getElementById('moveConfirmButton').addEventListener('click', () => {
            const dest = document.getElementById('moveDestination').value;
            const total = items.length;
            let done = 0;
            showProgressBar('move');

            items.forEach(({ fileName, itemType }) => {
                fetch(`/file-manager/move?item_name=${encodeURIComponent(fileName)}&path_param=${encodeURIComponent(FM.pathParam)}&item_type=${encodeURIComponent(itemType)}&destination_path=${encodeURIComponent(dest)}`, {
                    method: 'POST', headers: { 'X-CSRF-Token': csrf_token }
                })
                .then(r => r.json())
                .then(data => {
                    if (data.success) {
                        done++;
                        updateProgressBar('move', done, total);
                        if (done === total) finishProgressBar('move', 'Move complete!', 'Move completed successfully! Reloading page.');
                    } else {
                        showToast(data.error || 'Unknown error', 'error');
                    }
                })
                .catch(e => showToast(e.message, 'error'));
            });
        });
    });

    // DELETE
    document.getElementById('deleteButton').addEventListener('click', () => {
        const items = getSelectedItems();
        populateDrawerList('#deleteDrawer .selected-items-list', items);

        const confirmBtn = document.getElementById('deleteConfirmButton');
        const newBtn = confirmBtn.cloneNode(true);
        confirmBtn.parentNode.replaceChild(newBtn, confirmBtn);

        newBtn.addEventListener('click', () => {
            const total = items.length;
            let success = 0, failed = 0;
            showProgressBar('delete');
            showToast(`Deleting ${total} items...`, 'loading');
            const permanentCheckbox = document.getElementById('permanent-delete-checkbox');
            const mode = (!permanentCheckbox || permanentCheckbox.checked) ? 'permanent' : 'trash';

            items.forEach(({ fileName, itemType }) => {
                fetch(`/file-manager/delete?filename=${encodeURIComponent(fileName)}&path_param=${encodeURIComponent(FM.pathParam)}&item_type=${encodeURIComponent(itemType)}&mode=${mode}`, {
                    method: 'DELETE', headers: { 'X-CSRF-Token': csrf_token }
                })
                .then(r => r.json())
                .then(data => {
                    if (data.success) success++; else { failed++; showToast(`Error deleting "${fileName}": ${data.error}`, 'error'); }
                    updateProgressBar('delete', success + failed, total);
                    if (success + failed === total) {
                        showToast(failed > 0 ? 'Some items could not be deleted.' : 'All items deleted successfully.', failed > 0 ? 'warning' : 'success');
                        finishProgressBar('delete', 'Deletion complete', null);
                    }
                })
                .catch(e => { failed++; showToast(e.message, 'error'); });
            });
        });
    });

    // DOWNLOAD
    document.getElementById('downloadButton').addEventListener('click', () => {
        const row  = document.querySelector('#filemanager_table tbody tr.selected-row');
        const cell = row.querySelector('td[data-download-url]');
        if (row.dataset.type !== 'directory' && cell?.dataset.downloadUrl) {
            showToast(`Download for ${row.dataset.file} started.`, 'loading');
            window.location.href = cell.dataset.downloadUrl;
        }
    });

    // VIEW FILE
    document.getElementById('viewButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        let name = row.dataset.file;
        if (name === 'wp-config.php') name = 'wp_temp_openpanel-config.php';
        window.open(`/file-manager/view-file/${encodeURIComponent(FM.pathParam + '/')}${encodeURIComponent(name)}`, '_blank');
    });

    // EDIT FILE
    document.getElementById('editButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        let name = row.dataset.file;
        if (name === 'wp-config.php') name = 'wp_temp_openpanel-config.php';
        window.open(`/file-manager/edit-file/${encodeURIComponent(FM.pathParam + '/')}${encodeURIComponent(name)}`, '_blank');
    });

    // RENAME
    document.getElementById('renameButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        document.getElementById('renameInput').value = row.dataset.file;
        document.getElementById('old_name').value    = row.dataset.file;
    });

    // PERMISSIONS
    document.getElementById('permButton').addEventListener('click', () => {
        const items = getSelectedItems();
        const inputsContainer = document.getElementById('permissionsFilenameInputs');
        inputsContainer.innerHTML = items
            .map(({ fileName }) => `<input type="hidden" name="filename" value="${fileName.replace(/"/g, '&quot;')}">`)
            .join('');

        const listEl = document.querySelector('#permDrawer .selected-items-list');

        if (items.length === 1) {
            const row = document.querySelector('#filemanager_table tbody tr.selected-row');
            document.getElementById('permFilename').textContent = `${FM.i18n.changePermsFor} ${row.dataset.file}`;
            document.getElementById('c-oct').value = symbolicToOctal(row.dataset.permissions);
            listEl.innerHTML = '';
        } else {
            document.getElementById('permFilename').textContent =
                `${FM.i18n.changePermsFor} ${items.length} ${FM.i18n.selected}`;
            document.getElementById('c-oct').value = '';
            listEl.innerHTML = items.map(({ fileName }) => `<li>${fileName}</li>`).join('');
        }
    });

    // COMPRESS
    document.getElementById('compressButton').addEventListener('click', () => {
        const items = getSelectedItems();
        populateDrawerList('#compressDrawer .selected-items-list', items);

        document.getElementById('compressConfirmButton').addEventListener('click', () => {
            const archiveName = document.getElementById('compressDestination').value;
            const extension   = document.getElementById('compressArchiveFormat').value;
            showToast(`Started compressing to ${archiveName}.${extension}`, 'loading');

            const params = new URLSearchParams();
            params.append('archiveName', archiveName);
            params.append('pathParam', FM.pathParam);
            params.append('extension', extension);
            items.forEach(({ fileName }) => params.append('selectedFiles[]', fileName));

            fetch('/file-manager/create-archive', {
                method: 'POST',
                headers: { 'X-CSRF-TOKEN': csrf_token, 'Content-Type': 'application/x-www-form-urlencoded' },
                body: params
            })
            .then(r => r.json())
            .then(data => {
                if (data.success) showToast('Archive created! Reloading...', 'success');
                else console.error('Error creating archive:', data.error);
            })
            .catch(e => console.error('Error creating archive:', e))
            .finally(() => setTimeout(() => location.reload(), 500));
        });
    });

    // EXTRACT
    document.getElementById('extractButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        document.getElementById('archiveName').value        = row.dataset.file;
        document.getElementById('extractDestination').value = FM.pathParam;
    });
}

// UTILITIES

function getSelectedItems() {
    return Array.from(document.querySelectorAll('.selected-row')).map(el => ({
        fileName: el.dataset.file,
        itemType: el.dataset.type
    }));
}

function populateDrawerList(selector, items) {
    const list = document.querySelector(selector);
    list.innerHTML = items.map(({ fileName }) => `<li>${fileName}</li>`).join('');
}

function showProgressBar(type) {
    document.getElementById(`${type}-progress-container`).style.display = 'block';
}

function updateProgressBar(type, done, total) {
    const pct = Math.round((done / total) * 100);
    const bar = document.getElementById(`${type}-progress-bar`);
    bar.style.width = pct + '%';
    bar.setAttribute('aria-valuenow', pct);
    document.getElementById(`${type}-progress-label`).textContent = pct + '%';
}

function finishProgressBar(type, text, toastMsg) {
    document.getElementById(`${type}-progress-text`).textContent = text;
    updateProgressBar(type, 1, 1);
    if (toastMsg) showToast(toastMsg, 'success');
    setTimeout(() => location.reload(), 1000);
}

function formatBytes(bytes) {
    if (isNaN(bytes)) return bytes;
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
        bytes /= 1024;
        i++;
    }
    return bytes.toFixed(2) + ' ' + units[i];
}

function symbolicToOctal(symbolic) {
    if (!symbolic || symbolic.length < 9) return '';
    const map = { r: 4, w: 2, x: 1, '-': 0 };
    return symbolic.slice(-9).match(/.{3}/g)
        .map(p => p.split('').reduce((s, c) => s + (map[c] || 0), 0))
        .join('');
}

// legacy for table
document.addEventListener('alpine:init', () => {
    Alpine.data('permissionCell', () => ({
        symbolic: '',
        numeric: '',
        display: '',
        cache: window.__permCache || (window.__permCache = {}),
        init(symbolicValue) {
            this.symbolic = symbolicValue;
            this.display  = symbolicValue;
        },
        showNumeric() {
            if (!this.cache[this.symbolic]) {
                const perms = this.symbolic.slice(1);
                let numeric = '';
                for (let i = 0; i < perms.length; i += 3) {
                    const segment = perms.substr(i, 3);
                    let value = 0;
                    if (segment[0] === 'r') value += 4;
                    if (segment[1] === 'w') value += 2;
                    if (segment[2] === 'x') value += 1;
                    numeric += value.toString();
                }
                this.cache[this.symbolic] = numeric;
            }
            this.numeric = this.cache[this.symbolic];
            this.display = this.numeric;
        },
        showSymbolic() {
            this.display = this.symbolic;
        }
    }));
});


// INIT
document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('span.file-size-bytes').forEach(span => {
        const text = span.textContent.trim();
        if (/^\d+$/.test(text)) {
            span.textContent = formatBytes(parseInt(text, 10));
        }
    });

    if (typeof FM === 'undefined') return;

    disableButtons();
    initTableSelection();
    initSelectAll();
    initDirectorySizeCalculate();
    initActionButtons();
});
