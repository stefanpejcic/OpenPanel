var selectedRows = [];
var sveselect = 0;
var wasSelecting = false;
var isSelecting = false;
var startRowIndex = null;
var endRowIndex = null;
var anchorRow = null;

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

// only a single selected row lets you select its name as text, more than one clears any highlight
function syncTextSelect() {
    const table = document.getElementById('filemanager_table');
    if (!table) return;
    const single = table.querySelectorAll('tbody tr.selected-row').length === 1;
    table.classList.toggle('fm-single-select', single);
    if (!single) window.getSelection()?.removeAllRanges();
}

function disableButtons() {
    syncTextSelect();
    setDisabledAll('#viewButton, #editButton, #renameButton, #copyButton, #moveButton, #downloadButton, #deleteButton, #permButton, #compressButton, #extractButton', true);
    setClassAll('#viewButton, #editButton, #renameButton, #copyButton, #moveButton, #permButton, #compressButton, #extractButton', FM.btn.base);
    setClassAll('#downloadButton', FM.btn.download);
    setClassAll('#deleteButton', FM.btn.red);
}

function enableDisableButtons() {
    syncTextSelect();
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

// a fade-out still pending when the element fades back in would hide it again
function fadeIn(el, duration) {
    clearTimeout(el._fadeTimer);
    el.classList.remove('hidden');
    el.style.transition = `opacity ${duration}ms`;
    el.style.opacity = 0;
    requestAnimationFrame(() => el.style.opacity = 1);
}

function fadeOut(el, duration, callback) {
    el.style.transition = `opacity ${duration}ms`;
    el.style.opacity = 0;
    clearTimeout(el._fadeTimer);
    el._fadeTimer = setTimeout(() => { if (callback) callback(); }, duration);
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
        const t = new Date(`${mon} ${day} ${currentYear} ${hour}:${min}`).getTime();
        // ls only shows a time for the last 6 months, so a date days ahead is last year's, the slack covers server/browser timezones
        if (t > Date.now() + 2 * 86400000) {
            return new Date(`${mon} ${day} ${currentYear - 1} ${hour}:${min}`).getTime();
        }
        return t;
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
    icon.className = 'sort-icon ml-1 text-xs op-muted';
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

            // shift+click selects every row between the anchor and this one, ctrl/cmd+shift adds the range to the selection
            if (event.shiftKey && anchorRow && anchorRow.isConnected) {
                const rows = Array.from(table.querySelectorAll('tbody tr.clickable-row'));
                const from = rows.indexOf(anchorRow), to = rows.indexOf(this);
                if (!event.ctrlKey && !event.metaKey) {
                    table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                        r.classList.remove('selected-row')
                    );
                    selectedRows = [];
                }
                rows.slice(Math.min(from, to), Math.max(from, to) + 1).forEach(r => {
                    r.classList.add('selected-row');
                    if (!selectedRows.includes(r.dataset.rowIndex)) selectedRows.push(r.dataset.rowIndex);
                });
                sveselect = 0;
                setSelectAllLabel(false);
                enableDisableButtons();
                updateSelectedOptionsDisplay();
                return;
            }

            anchorRow = this;
            if (event.ctrlKey || event.metaKey) {
                toggleClasses(this, 'selected-row');
                const idx = selectedRows.indexOf(rowIndex);
                if (idx === -1) selectedRows.push(rowIndex);
                else selectedRows.splice(idx, 1);
                sveselect = 0;
                setSelectAllLabel(false);
            } else {
                table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                    r.classList.remove('selected-row')
                );
                toggleClasses(this, 'selected-row');
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
        if (event.ctrlKey || event.metaKey || event.shiftKey || event.button !== 0) return;

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
                    row.classList.add('selected-row');
                    if (!selectedRows.includes(rowIndex)) selectedRows.push(rowIndex);
                } else {
                    row.classList.remove('selected-row');
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

    // click outside to deselect, a target its own handler already removed (the permissions sheet's Cancel) doesn't count
    document.addEventListener('click', event => {
        const t = event.target;
        if (
            t.isConnected &&
            !t.closest('#filemanager_table') &&
            !t.closest('#mainButtons') &&
            !t.closest('#selectedOptions') &&
            !t.closest('#fmContextMenu') &&
            !t.closest('.modal.fade') &&
            !t.closest('.btn-group') &&
            !t.closest('.fmdrawer') &&
            !t.closest('#fmPickerModal') &&
            !wasSelecting
        ) {
            table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                r.classList.remove('selected-row')
            );
            selectedRows = [];
            anchorRow = null;
            sveselect = 0;
            setSelectAllLabel(false);
            disableButtons();
            updateSelectedOptionsDisplay();
        }
        wasSelecting = false;
    });
}

// RIGHT-CLICK MENU (classic view only, mainButtons only exists in that view)

function initContextMenu() {
    const menu = document.getElementById('fmContextMenu');
    const mainButtons = document.getElementById('mainButtons');
    if (!menu || !mainButtons) return;

    const actionIds = ['viewButton', 'editButton', 'downloadButton', 'renameButton', 'copyButton', 'moveButton', 'compressButton', 'extractButton', 'permButton', 'deleteButton'];

    menu.innerHTML = actionIds.map(id => {
        const btn = document.getElementById(id);
        if (!btn) return '';
        const icon = btn.querySelector('i')?.className || '';
        const label = btn.querySelector('span')?.textContent.trim() || '';
        const danger = id === 'deleteButton' ? ' is-danger' : '';
        return `<li data-target="${id}" class="flex items-center gap-x-2 px-3 py-1.5 text-sm cursor-pointer op-hover${danger}"><i class="${icon}"></i>${label}</li>`;
    }).join('');

    const table = document.getElementById('filemanager_table');

    table.addEventListener('contextmenu', event => {
        const row = event.target.closest('tr.clickable-row');
        if (!row) return;
        event.preventDefault();

        if (!row.classList.contains('selected-row')) {
            table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                r.classList.remove('selected-row')
            );
            row.classList.add('selected-row');
            selectedRows = [row.dataset.rowIndex];
            sveselect = 0;
            setSelectAllLabel(false);
        }

        enableDisableButtons();
        updateSelectedOptionsDisplay();

        menu.querySelectorAll('li[data-target]').forEach(li => {
            const btn = document.getElementById(li.dataset.target);
            li.classList.toggle('hidden', !btn || btn.disabled);
        });

        showContextMenu(event.pageX, event.pageY);
    });

    menu.addEventListener('click', event => {
        const li = event.target.closest('li[data-target]');
        if (!li || li.classList.contains('hidden')) return;
        hideContextMenu();
        document.getElementById(li.dataset.target)?.click();
    });

    document.addEventListener('click', () => hideContextMenu());
    document.addEventListener('scroll', () => hideContextMenu(), true);
    window.addEventListener('resize', () => hideContextMenu());
    document.addEventListener('keydown', event => {
        if (event.key === 'Escape') hideContextMenu();
    });
}

function showContextMenu(x, y) {
    const menu = document.getElementById('fmContextMenu');
    menu.classList.remove('hidden');
    const maxX = window.scrollX + document.documentElement.clientWidth - menu.offsetWidth - 8;
    const maxY = window.scrollY + document.documentElement.clientHeight - menu.offsetHeight - 8;
    menu.style.left = Math.max(window.scrollX, Math.min(x, maxX)) + 'px';
    menu.style.top  = Math.max(window.scrollY, Math.min(y, maxY)) + 'px';
}

function hideContextMenu() {
    document.getElementById('fmContextMenu')?.classList.add('hidden');
}

// SELECT ALL BUTTON

function initSelectAll() {
    document.getElementById('SelectAll-button').addEventListener('click', () => {
        const table = document.getElementById('filemanager_table');
        const rows  = table.querySelectorAll('tbody tr.clickable-row');

        if (!sveselect) {
            rows.forEach(r => r.classList.add('selected-row'));
            selectedRows = Array.from(rows).map((_, i) => String(i));
            sveselect = 1;
            setSelectAllLabel(true);
        } else {
            rows.forEach(r => r.classList.remove('selected-row'));
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
            const folder = FM.pathParam ? FM.pathParam + '/' + folderName : folderName;
            const url = `/json/directory-size?folder=${encodeURIComponent(folder)}`;

            fetch(url, { headers: { 'X-CSRF-Token': csrf_token } })
                .then(r => r.json())
                .then(data => { cell.textContent = data.size; })
                .catch(error => {
                    showToast(`${FM.i18n.folderSizeError} ${error.message || error}`, 'error');
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
    // COPY / MOVE / COMPRESS / EXTRACT all go through the folder picker modal
    document.getElementById('copyButton').addEventListener('click', () => openPicker('copy', getSelectedItems()));
    document.getElementById('moveButton').addEventListener('click', () => openPicker('move', getSelectedItems()));
    document.getElementById('compressButton').addEventListener('click', () => openPicker('compress', getSelectedItems()));
    document.getElementById('extractButton').addEventListener('click', () => openPicker('extract', getSelectedItems()));

    // DELETE (inline, first selected row gets the buttons, every selected name is crossed out)
    document.getElementById('deleteButton').addEventListener('click', () => {
        const items = getSelectedItems();
        const rows = Array.from(document.querySelectorAll('#filemanager_table tbody tr.selected-row'));
        if (!rows.length) return;

        const crossed = rows.slice(1).flatMap(r => Array.from(r.querySelectorAll('td:first-child .op-row-title b')));
        const crossOut = on => crossed.forEach(el => {
            el.style.textDecoration = on ? 'line-through' : '';
            el.style.color = on ? '#dc2626' : '';
        });

        const firstCell = rows[0].querySelector('td');
        const icon = firstCell.querySelector('i')?.outerHTML || '';
        const nameEl = document.createElement('span');
        nameEl.textContent = rows[0].dataset.file;

        inlineEdit(firstCell, `
            <div class="w-4 text-center">${icon}</div>
            <span class="text-sm is-danger" style="text-decoration: line-through;">${nameEl.innerHTML}</span>
        `, (input, editor) => {
            const permanent = !FM.canTrash || editor.querySelector('.fm-inline-skip-trash').checked;
            const total = items.length;
            let failed = 0, done = 0;
            const saveBtn = editor.querySelector('[data-act="save"]');
            editor.querySelectorAll('button, input').forEach(b => b.disabled = true);

            items.forEach(({ fileName, itemType }) => {
                fetch(`/file-manager/delete?filename=${encodeURIComponent(fileName)}&path_param=${encodeURIComponent(FM.pathParam)}&item_type=${encodeURIComponent(itemType)}&mode=${permanent ? 'permanent' : 'trash'}`, {
                    method: 'DELETE', headers: { 'X-CSRF-Token': csrf_token }
                })
                .then(r => r.json())
                .then(data => { if (!data.success) throw new Error(data.error); })
                .catch(e => { failed++; showToast(`${FM.i18n.deleteError} "${fileName}": ${e.message}`, 'error'); })
                .finally(() => {
                    done++;
                    if (total > 1) saveBtn.innerHTML = `<i class="bi bi-trash"></i> ${done}/${total}`;
                    if (done < total) return;
                    if (failed) {
                        showToast(failed === total ? FM.i18n.deleteFailed : FM.i18n.deletePartial, 'warning');
                    } else {
                        showToast(permanent ? FM.i18n.deleted : FM.i18n.movedToTrash, 'success');
                    }
                    setTimeout(() => location.reload(), failed ? 1500 : 500);
                });
            });
        }, {
            saveLabel: items.length > 1 ? `${FM.i18n.delete} (${items.length})` : FM.i18n.delete,
            saveIcon: 'bi-trash',
            saveClass: 'op-btn-danger',
            extra: FM.canTrash ? `<label class="inline-flex items-center gap-1.5 text-sm cursor-pointer op-muted"><input type="checkbox" class="fm-inline-skip-trash op-native-check"> ${FM.i18n.skipTrash}</label>` : '',
            onClose: () => crossOut(false),
        });
        crossOut(true);
    });

    // DOWNLOAD
    document.getElementById('downloadButton').addEventListener('click', () => {
        const row  = document.querySelector('#filemanager_table tbody tr.selected-row');
        const cell = row.querySelector('td[data-download-url]');
        if (row.dataset.type !== 'directory' && cell?.dataset.downloadUrl) {
            showToast(FM.i18n.downloadStarted.replace('{file}', row.dataset.file), 'loading');
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

    // RENAME (inline in the name cell)
    document.getElementById('renameButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        if (!row) return;
        const nameCell = row.querySelector('td');
        const icon = nameCell.querySelector('i')?.outerHTML || '';
        const name = row.dataset.file;

        const input = inlineEdit(nameCell, `
            <div class="w-4 text-center">${icon}</div>
            <input type="text" class="fm-inline-input select-text block w-64 appearance-none op-input op-input-sm">
        `, input => {
            const newName = input.value.trim();
            if (!newName) return false;
            if (newName === name) return true;
            postForm('/file-manager/rename', [['path_param', FM.pathParam], ['old_name', name], ['new_name', newName]]);
        });
        input.value = name;
        // preselect the name without its extension, like most file managers do
        const dot = row.dataset.type === 'file' ? name.lastIndexOf('.') : -1;
        input.setSelectionRange(0, dot > 0 ? dot : name.length);
    });

    // PERMISSIONS: the toolbar/selection bar button and the row menu edit the selection, a mode in the Permissions column edits its own row
    document.getElementById('permButton').addEventListener('click', () => {
        openPermSheet(Array.from(document.querySelectorAll('#filemanager_table tbody tr.selected-row')));
    });
    // runs after the row's own click handler has selected it
    document.getElementById('filemanager_table').addEventListener('click', e => {
        const btn = e.target.closest('button[data-fm-perm]');
        if (btn) openPermSheet([btn.closest('tr')]);
    });
}

// PERMISSIONS EDITOR (fm_perm_sheet in filemanager_partials.html)

const rwx = m => m.split('').map(c => { const n = +c; return (n & 4 ? 'r' : '-') + (n & 2 ? 'w' : '-') + (n & 1 ? 'x' : '-'); }).join('');
const fillIn = (s, v) => String(s || '').replace(/\{(\w+)\}/g, (m, k) => (k in v ? v[k] : m));

// symlinks are left out, chmod on one changes whatever it points to
function openPermSheet(rows) {
    const items = rows.filter(r => r && r.dataset.mode && r.dataset.type !== 'symlink').map(r => ({
        name: r.dataset.file,
        isDir: r.dataset.type === 'directory',
        mode: r.dataset.mode,
        sym: (r.dataset.permissions || '').slice(0, 10),
        owner: r.dataset.owner || '',
        group: r.dataset.group || '',
    }));
    if (!items.length) {
        if (rows.length) showToast(escHTML(window.FM_PERM?.i18n.symlinksOnly || ''), 'error');
        return;
    }
    window.dispatchEvent(new CustomEvent('fm-perm-open', { detail: items }));
}

// the changed rows get the new mode in place, same markup as files_table renders
function updatePermRow(name, m) {
    const t = window.FM_PERM?.i18n || {};
    document.querySelectorAll('#filemanager_table tbody tr[data-file]').forEach(row => {
        if (row.dataset.file !== name) return;
        const old = row.dataset.permissions || '';
        const sym = (old[0] || '-') + rwx(m);
        row.dataset.permissions = sym + old.slice(10);
        row.dataset.mode = m;
        const btn = row.querySelector('button[data-fm-perm]');
        if (!btn) return;
        btn.querySelector('.op-perm').textContent = '0' + m;
        btn.querySelector('.fm-perm-sym').textContent = sym;
        btn.setAttribute('aria-label', fillIn(t.cellLabel, { mode: '0' + m, name }));
        btn.classList.toggle('is-risky', (+m[2] & 2) > 0);
        btn.classList.add('is-new');
        setTimeout(() => btn.classList.remove('is-new'), 2400);
    });
}

// one mode for every picked item, posted to /file-manager/permissions like the old inline editor but read back as JSON
function fmPermSheet() {
    const t = window.FM_PERM?.i18n || { presets: [] };
    return {
        t,
        pm: null,
        busy: false,
        failed: '',
        opener: null,
        // the x-if body can still re-evaluate once while the sheet closes, so nothing below assumes pm is set
        items() { return this.pm ? this.pm.items : []; },
        value() { return this.pm ? this.pm.value.trim() : ''; },
        open(items) {
            if (!items || !items.length) return;
            const modes = [...new Set(items.map(i => i.mode))];
            this.opener = document.activeElement;
            this.busy = false;
            this.failed = '';
            this.pm = { items, value: modes.length === 1 ? modes[0] : '', recursive: false, dirValue: '', dirTouched: false };
            this.$nextTick(() => { const el = document.getElementById('fm-pm-octal'); if (el) { el.focus(); el.select(); } });
        },
        close() {
            if (this.busy) return;
            this.pm = null;
            if (this.opener && this.opener.isConnected) this.opener.focus();
            this.opener = null;
        },
        // the server's rule (crud.go permissionsRE): three digits, a leading 0 is fine, setuid/setgid/sticky refused
        valid() { return /^0?[0-7]{3}$/.test(this.value()); },
        mode() { return this.value().slice(-3); },
        error() {
            const v = this.value();
            if (!v || this.valid()) return '';
            return /^[1-7][0-7]{3}$/.test(v) ? t.errSpecial : t.errFormat;
        },
        hint() { return this.error() || (this.value() ? t.hint : t.hintMixed); },
        bit(k, b) { return this.valid() && (+this.mode()[k] & b) > 0; },
        setBit(k, b, on) {
            const d = (this.valid() ? this.mode() : '000').split('').map(Number);
            d[k] = on ? d[k] | b : d[k] & ~b;
            this.pm.value = d.join('');
        },
        bitLabel(w, b) { return w.n + ' ' + (b === 4 ? t.read : b === 2 ? t.write : t.execute); },
        sym(m, dir) { return (dir ? 'd' : '-') + rwx(m); },
        dirs() { return this.items().filter(i => i.isDir); },
        files() { return this.items().filter(i => !i.isDir); },
        allDirs() { return this.files().length === 0; },
        who() {
            const first = this.items()[0] || {};
            return [{ k: 0, n: t.owner, s: first.owner }, { k: 1, n: t.group, s: first.group }, { k: 2, n: t.others, s: t.everyoneElse }];
        },
        names() {
            const n = this.items().map(i => i.name);
            if (n.length < 2) return '';
            return n.length > 4 ? fillIn(t.more, { names: n.slice(0, 4).join(', '), n: n.length - 4 }) : n.join(', ');
        },
        target() {
            const path = '/var/www/html' + (FM.pathParam ? '/' + FM.pathParam.replace(/^\/+|\/+$/g, '') : '');
            const all = this.items();
            if (all.length < 2) return all.length ? fillIn(t.inPath, { what: all[0].name, path }) : '';
            const d = this.dirs().length, f = all.length - d;
            const parts = [d && (d === 1 ? t.oneFolder : fillIn(t.folders, { n: d })), f && (f === 1 ? t.oneFile : fillIn(t.files, { n: f }))].filter(Boolean);
            return fillIn(t.inPath, { what: parts.length === 2 ? fillIn(t.and, { a: parts[0], b: parts[1] }) : parts[0], path });
        },
        current() {
            const all = this.items(), modes = [...new Set(all.map(i => i.mode))];
            if (!modes.length) return { oct: '', sym: '' };
            if (modes.length > 1) return { oct: t.mixed, sym: modes.map(m => '0' + m).join(', ') };
            const one = all.length === 1 ? all[0] : null;
            return { oct: '0' + modes[0], sym: one && one.sym ? one.sym : this.sym(modes[0], this.allDirs()) };
        },
        suggested(p) { return !!p.for && (p.for === 'dir' ? this.files().length === 0 : this.dirs().length === 0); },
        recursive() { return !!this.pm && this.pm.recursive && this.dirs().length > 0; },
        // folders inside need execute to be opened: every digit with read gets execute too (644 -> 755, 640 -> 750)
        folderFrom(m) { return m.split('').map(d => (+d & 4) ? String(+d | 1) : d).join(''); },
        dirValue() { return this.pm.dirTouched ? this.pm.dirValue.trim() : (this.valid() ? this.folderFrom(this.mode()) : ''); },
        dirValid() { return /^0?[0-7]{3}$/.test(this.dirValue()); },
        dirMode() { return this.dirValue().slice(-3); },
        dirError() {
            const v = this.dirValue();
            if (!v || this.dirValid()) return '';
            return /^[1-7][0-7]{3}$/.test(v) ? t.errSpecial : t.errFormat;
        },
        setDir(v) { this.pm.dirValue = v; this.pm.dirTouched = true; },
        sameForFolders() { this.setDir(this.mode()); },
        ready() { return this.valid() && (!this.recursive() || this.dirValid()); },
        // what the selection itself gets: a folder applied recursively takes the folder mode
        newMode() { return this.recursive() && this.allDirs() && this.dirValid() ? this.dirMode() : this.mode(); },
        unchanged() { return this.valid() && !this.recursive() && this.items().every(i => i.mode === this.mode()); },
        warnings() {
            if (!this.valid()) return [];
            const m = this.mode(), o = +m[0], g = +m[1], w = [];
            const rec = this.recursive(), many = rec || this.files().length > 1;
            const files = this.files().length > 0 || rec, dirs = this.dirs().length > 0;
            if (+m[2] & 2) w.push({ tone: 'danger', text: fillIn(t.warnWorld, { mode: '0' + m }) });
            if (files && (o | g | +m[2]) & 1) w.push({ tone: 'warning', text: (rec ? t.warnExecInside : many ? t.warnExecMany : t.warnExecOne) + ' ' + t.warnExecWhy });
            const dm = rec && this.dirValid() ? this.dirMode() : m;
            if (dirs && (+dm[0] & 5) !== 5) w.push({ tone: 'danger', text: rec ? t.warnDirInside : t.warnDir });
            if (files && !(o & 4)) w.push({ tone: 'danger', text: many ? t.warnReadMany : t.warnReadOne });
            return w;
        },
        applyLabel() {
            if (this.busy) return t.applying;
            if (!this.valid()) return t.apply;
            if (this.recursive() && this.dirValid() && this.dirMode() !== this.mode()) return fillIn(t.applyBoth, { files: '0' + this.mode(), folders: '0' + this.dirMode() });
            return fillIn(t.applyMode, { mode: '0' + this.mode() });
        },
        async apply() {
            if (!this.pm || !this.ready() || this.busy) return;
            const p = this.pm, m = this.mode(), rec = this.recursive(), dm = rec ? this.dirMode() : m;
            if (this.unchanged()) {
                this.close();
                showToast(escHTML(fillIn(t.noChange, { mode: '0' + m })), 'success');
                return;
            }
            const body = new URLSearchParams({ csrf_token, path_param: FM.pathParam, permissions: m });
            p.items.forEach(i => body.append('filename', i.name));
            if (rec) {
                body.append('recursive', 'on');
                body.append('folder_permissions', dm);
            }
            this.busy = true;
            this.failed = '';
            let res = null;
            try {
                const r = await fetch('/file-manager/permissions', { method: 'POST', headers: { 'Accept': 'application/json', 'X-CSRF-Token': csrf_token }, body });
                res = await r.json().catch(() => null);
                if (!res) throw new Error('HTTP ' + r.status);
                if (!r.ok) throw new Error(res.error || 'HTTP ' + r.status);
            } catch (e) {
                // shown in the sheet, a toast would sit under its scrim
                this.busy = false;
                this.failed = e.message || FM.i18n.unknownError;
                return;
            }
            this.busy = false;
            const changed = res.changed || [];
            const isDir = new Set(p.items.filter(i => i.isDir).map(i => i.name));
            changed.forEach(name => updatePermRow(name, isDir.has(name) ? dm : m));
            // when nothing went through the sheet stays open with the reason
            if (!changed.length) {
                this.failed = res.error || FM.i18n.unknownError;
                return;
            }
            this.close();
            const inside = rec && p.items.some(i => i.isDir && changed.includes(i.name));
            const msg = changed.length === 1
                ? fillIn(inside ? t.doneOneInside : t.doneOne, { mode: '0' + m })
                : fillIn(inside ? t.doneManyInside : t.doneMany, { n: changed.length, mode: '0' + m });
            showToast(escHTML(msg), 'success');
            // some went through and some didn't
            if (res.error) showToast(escHTML(res.error), 'error');
        },
    };
}

// FOLDER PICKER MODAL (copy, move, compress, extract)

function escHTML(str) {
    const d = document.createElement('div');
    d.textContent = str;
    return d.innerHTML;
}

const joinPath = (...parts) => parts.filter(Boolean).join('/').replace(/\/+/g, '/').replace(/^\/|\/$/g, '');

function openPicker(action, items) {
    if (!items.length) return;
    const $ = id => document.getElementById(id);
    const modal = $('fmPickerModal');
    const list = $('fmPickerList');
    const destInput = $('fmPickerDest');
    const confirmBtn = $('fmPickerConfirm');
    const status = $('fmPickerStatus');
    const nameInput = $('fmPickerArchiveName');
    const subfolder = $('fmPickerSubfolder');

    // moving/copying a folder into itself (or below it) makes no sense, so those rows are greyed out
    const blocked = (action === 'copy' || action === 'move')
        ? items.filter(it => it.itemType === 'directory').map(it => joinPath(FM.pathParam, it.fileName))
        : [];
    const isBlocked = p => blocked.some(b => p === b || p.startsWith(b + '/'));

    let cur = FM.pathParam.replace(/^\/|\/$/g, '');
    let busy = false;

    const archiveBase = items[0].fileName.replace(/\.(tar\.gz|tgz|tar|zip|gz)$/i, '');
    $('fmPickerTitle').textContent = items.length > 1
        ? `${FM.i18n[action]} ${items.length} ${FM.i18n.items}`
        : `${FM.i18n[action]} ${items[0].fileName}`;
    $('fmPickerItems').textContent = items.length > 1 ? items.map(it => it.fileName).join(', ') : '';
    $('fmPickerNameRow').classList.toggle('hidden', action !== 'compress');
    $('fmPickerSubfolderRow').classList.toggle('hidden', action !== 'extract');
    $('fmPickerSubfolderName').textContent = archiveBase + '/';
    subfolder.checked = false;
    confirmBtn.textContent = { copy: FM.i18n.copyHere, move: FM.i18n.moveHere, compress: FM.i18n.compress, extract: FM.i18n.extractHere }[action];
    confirmBtn.disabled = false;
    status.className = 'hidden';
    $('fmPickerCancel').textContent = FM.i18n.cancel;

    if (action === 'compress') {
        const folderName = FM.pathParam.split('/').filter(Boolean).pop();
        nameInput.value = items.length === 1 ? items[0].fileName.replace(/\.[^.]+$/, '') || items[0].fileName : (folderName || 'archive');
        nameInput.style.borderColor = '';
    }

    function setDest(path) {
        cur = path;
        destInput.value = '/' + cur;
        destInput.style.borderColor = '';
        // moving into the folder the items already live in is a no-op
        confirmBtn.disabled = (action === 'move' || action === 'copy') && (cur === FM.pathParam.replace(/^\/|\/$/g, '') || isBlocked(cur));
    }

    function renderCrumbs() {
        const parts = cur ? cur.split('/') : [];
        const crumbs = [`<button type="button" data-path="" class="cursor-pointer op-muted"><i class="bi bi-house"></i></button>`];
        parts.forEach((part, i) => {
            crumbs.push(`<span class="op-muted">/</span><button type="button" data-path="${escHTML(parts.slice(0, i + 1).join('/'))}" class="cursor-pointer truncate op-muted">${escHTML(part)}</button>`);
        });
        $('fmPickerCrumbs').innerHTML = crumbs.join('');
    }

    async function load(path) {
        list.innerHTML = `<li class="px-3 py-2 op-muted">${FM.i18n.working}</li>`;
        let folders;
        try {
            const res = await fetch('/json/folders/' + path.split('/').map(encodeURIComponent).join('/'));
            if (!res.ok) throw new Error();
            folders = (await res.json()).folders || [];
            // the malware scanner's quarantine is not a place to put files
            if (!path) folders = folders.filter(f => f.name !== '.quarantine');
        } catch (e) {
            destInput.style.borderColor = '#ef4444';
            showToast(FM.i18n.folderNotFound, 'error');
            if (path !== cur) return load(cur);
            folders = [];
        }

        setDest(path);
        renderCrumbs();
        const rows = [];
        if (cur) {
            rows.push(`<li><button type="button" data-path="${escHTML(cur.split('/').slice(0, -1).join('/'))}" class="flex w-full items-center gap-2 px-3 py-2 text-left cursor-pointer op-hover"><i class="bi bi-arrow-up"></i> ..</button></li>`);
        }
        folders.sort((x, y) => x.name.localeCompare(y.name)).forEach(f => {
            const p = joinPath(cur, f.name);
            const off = isBlocked(p);
            rows.push(`<li><button type="button" data-path="${escHTML(p)}" ${off ? 'disabled' : ''} class="flex w-full items-center gap-2 px-3 py-2 text-left ${off ? 'cursor-not-allowed op-muted' : 'op-hover cursor-pointer'}"><i class="bi bi-folder-fill is-warning"></i><span class="truncate min-w-0">${escHTML(f.name)}</span>${f.has_subfolders && !off ? '<i class="bi bi-chevron-right ml-auto op-muted"></i>' : ''}</button></li>`);
        });
        if (!folders.length) rows.push(`<li class="px-3 py-2 op-muted">${FM.i18n.noSubfolders}</li>`);
        list.innerHTML = rows.join('');
    }

    function showStatus(kind, html) {
        status.className = 'op-alert' + (kind === 'error' ? ' is-danger' : '');
        status.innerHTML = html;
    }

    function close() {
        if (busy) return;
        modal.classList.add('hidden');
        document.removeEventListener('keydown', onKey);
        modal.onclick = null;
        list.onclick = $('fmPickerCrumbs').onclick = destInput.onkeydown = confirmBtn.onclick = $('fmPickerNewFolder').onclick = null;
    }

    function onKey(e) {
        if (e.key === 'Escape') close();
    }

    function newFolderRow() {
        if (list.querySelector('.fm-picker-new')) return;
        const li = document.createElement('li');
        li.className = 'fm-picker-new flex items-center gap-2 px-3 py-1.5';
        li.innerHTML = `<i class="bi bi-folder-fill is-warning"></i><input type="text" data-no-live-validate class="select-text block w-full appearance-none op-input op-input-sm">`;
        list.prepend(li);
        const input = li.querySelector('input');
        input.focus();
        input.addEventListener('keydown', async e => {
            e.stopPropagation();
            if (e.key === 'Escape') return li.remove();
            if (e.key !== 'Enter') return;
            const name = input.value.trim();
            if (!name || name.includes('/')) { input.style.borderColor = '#ef4444'; return; }
            const body = new URLSearchParams({ csrf_token, path_param: cur, foldername: name });
            await fetch('/file-manager/new-directory', { method: 'POST', body }).catch(() => {});
            load(joinPath(cur, name));
        });
    }

    async function runAction() {
        const dest = destInput.value.trim().replace(/^\/+|\/+$/g, '');
        busy = true;
        confirmBtn.disabled = true;
        modal.querySelectorAll('[data-picker-close]').forEach(b => b.disabled = true);
        const errors = [];

        if (action === 'copy' || action === 'move') {
            let done = 0;
            showStatus('info', `${FM.i18n.working} 0/${items.length}`);
            await Promise.all(items.map(({ fileName, itemType }) =>
                fetch(`/file-manager/${action}?item_name=${encodeURIComponent(fileName)}&path_param=${encodeURIComponent(FM.pathParam)}&item_type=${encodeURIComponent(itemType)}&destination_path=${encodeURIComponent(dest)}`, {
                    method: 'POST', headers: { 'X-CSRF-Token': csrf_token }
                })
                .then(r => r.json())
                .then(data => { if (!data.success) throw new Error(data.error || FM.i18n.unknownError); })
                .catch(e => errors.push(`${fileName}: ${e.message}`))
                .finally(() => { done++; showStatus('info', `${FM.i18n.working} ${done}/${items.length}`); })
            ));
        } else {
            // compress and extract answer with a redirect + flash, so just wait and reload to show it
            const body = new URLSearchParams({ csrf_token });
            if (action === 'compress') {
                body.append('archiveName', joinPath(dest, nameInput.value.trim()));
                body.append('extension', $('fmPickerExt').value);
                body.append('pathParam', FM.pathParam);
                items.forEach(({ fileName }) => body.append('selectedFiles[]', fileName));
            } else {
                body.append('archiveName', items[0].fileName);
                body.append('path_param', FM.pathParam);
                body.append('extractDestination', '/' + (subfolder.checked ? joinPath(dest, archiveBase) : dest));
            }
            showStatus('info', FM.i18n.working);
            await fetch(action === 'compress' ? '/file-manager/create-archive' : '/file-manager/extract-archive', {
                method: 'POST', headers: { 'X-CSRF-Token': csrf_token }, body
            })
            .then(async r => { if (!r.ok) throw new Error((await r.text()).trim() || 'HTTP ' + r.status); })
            .catch(e => errors.push(e.message));
        }

        busy = false;
        if (!errors.length) {
            showStatus('info', FM.i18n.doneReloading);
            setTimeout(() => location.reload(), 400);
            return;
        }
        modal.querySelectorAll('[data-picker-close]').forEach(b => b.disabled = false);
        showStatus('error', `<ul class="list-disc pl-5">${errors.map(e => `<li>${escHTML(e)}</li>`).join('')}</ul>`);
        // some items may have gone through, so closing should show the new state
        $('fmPickerCancel').textContent = FM.i18n.close;
        $('fmPickerCancel').onclick = () => location.reload();
    }

    list.onclick = $('fmPickerCrumbs').onclick = e => {
        const btn = e.target.closest('button[data-path]');
        if (btn && !btn.disabled) load(btn.dataset.path);
    };
    destInput.onkeydown = e => {
        if (e.key === 'Enter') { e.preventDefault(); load(destInput.value.trim().replace(/^\/+|\/+$/g, '')); }
    };
    destInput.oninput = () => {
        const typed = destInput.value.trim().replace(/^\/+|\/+$/g, '');
        confirmBtn.disabled = (action === 'move' || action === 'copy') && (typed === FM.pathParam.replace(/^\/|\/$/g, '') || isBlocked(typed));
    };
    $('fmPickerNewFolder').onclick = newFolderRow;
    confirmBtn.onclick = () => {
        if (action === 'compress' && (!nameInput.value.trim() || nameInput.value.includes('/'))) {
            nameInput.style.borderColor = '#ef4444';
            return nameInput.focus();
        }
        runAction();
    };
    nameInput.oninput = () => { nameInput.style.borderColor = ''; };
    nameInput.onkeydown = e => { if (e.key === 'Enter') { e.preventDefault(); confirmBtn.click(); } };
    modal.querySelectorAll('[data-picker-close]').forEach(b => { b.onclick = close; });
    modal.onclick = e => { if (e.target === modal) close(); };
    document.addEventListener('keydown', onKey);

    modal.classList.remove('hidden');
    load(cur);
    if (action === 'compress') { nameInput.focus(); nameInput.select(); } else confirmBtn.focus();
}

// UTILITIES

// plain form POST so the server's redirect + flash work as usual
function postForm(action, fields) {
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = action;
    form.style.display = 'none';
    [['csrf_token', csrf_token], ...fields].forEach(([name, val]) => {
        const el = document.createElement('input');
        el.type = 'hidden';
        el.name = name;
        el.value = val;
        form.appendChild(el);
    });
    document.body.appendChild(form);
    form.submit();
}

function getSelectedItems() {
    return Array.from(document.querySelectorAll('.selected-row')).map(el => ({
        fileName: el.dataset.file,
        itemType: el.dataset.type
    }));
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

// swaps a table cell for an input + save/cancel, onSave returns false to flag the input red, true to just close
let closeInlineEdit = null;

function inlineEdit(cell, innerHTML, onSave, opts = {}) {
    if (closeInlineEdit) closeInlineEdit();

    const buttons = `
            ${opts.extra || ''}
            <button type="button" data-act="save" class="op-btn op-btn-sm ${opts.saveClass || 'op-btn-primary'}"><i class="bi ${opts.saveIcon || 'bi-check-lg'}"></i> ${opts.saveLabel || FM.i18n.save}</button>
            <button type="button" data-act="cancel" class="gap-1 op-btn op-btn-sm op-btn-secondary"><i class="bi bi-x-lg"></i> ${FM.i18n.cancel}</button>`;

    const editor = document.createElement('td');
    editor.className = cell.className.replace('text-center', '') + ' fm-inline-editor';
    // last column has no room to the right, so float the buttons left over the neighbouring cells
    editor.innerHTML = opts.buttonsBefore ? `
        <div class="flex items-center justify-center py-0.5" style="position: relative;">
            <div class="flex items-center gap-2 pl-2" style="position: absolute; right: 100%; top: 50%; transform: translateY(-50%); margin-right: .5rem; white-space: nowrap; z-index: 10;">${buttons}</div>
            ${innerHTML}
        </div>` : `
        <div class="flex flex-wrap items-center gap-2 py-0.5">
            ${innerHTML}
            ${buttons}
        </div>`;

    cell.style.display = 'none';
    cell.insertAdjacentElement('afterend', editor);
    const input = editor.querySelector('.fm-inline-input');
    const saveBtn = editor.querySelector('[data-act="save"]');

    function close() {
        if (opts.onClose) opts.onClose();
        editor.remove();
        cell.style.display = '';
        document.removeEventListener('click', onOutside, true);
        closeInlineEdit = null;
    }

    function save() {
        const result = onSave(input, editor);
        if (result === false && input) {
            input.style.borderColor = '#ef4444';
            input.focus();
        } else if (result === true) {
            close();
        }
    }

    function onOutside(e) {
        if (!editor.contains(e.target)) close();
    }

    // keep row selection / drag-select out of clicks inside the editor
    ['click', 'mousedown', 'dblclick', 'contextmenu'].forEach(ev => editor.addEventListener(ev, e => e.stopPropagation()));
    saveBtn.addEventListener('click', save);
    editor.querySelector('[data-act="cancel"]').addEventListener('click', close);
    if (input) {
        input.addEventListener('input', () => { input.style.borderColor = ''; });
        input.addEventListener('keydown', e => {
            if (e.key === 'Enter') { e.preventDefault(); save(); }
        });
    }
    editor.addEventListener('keydown', e => {
        if (e.key === 'Escape') { e.preventDefault(); close(); }
    });
    // defer so the click that opened the editor doesn't immediately close it
    setTimeout(() => document.addEventListener('click', onOutside, true), 0);

    closeInlineEdit = close;
    (input || saveBtn).focus();
    return input;
}

// the trash page's permissions cell, /files has its own in files_table
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
    initContextMenu();
});
