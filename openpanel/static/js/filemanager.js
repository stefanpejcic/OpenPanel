var selectedRows = [];
var sveselect = 0;
var wasSelecting = false;
var isSelecting = false;
var startRowIndex = null;
var endRowIndex = null;
var anchorRow = null;

// HELPERS

// classes a selected row gets, the row hover skips selected rows
const SELECTED = ['selected-row', 'bg-gray-50', 'dark:bg-gray-900'];

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
    const icon = deselect ? 'ri-cursor-line' : 'ri-cursor-line';
    const text = deselect ? FM.i18n.deselect : FM.i18n.selectAll;
    document.getElementById('spanAll').innerHTML = `<i class="${icon}"></i> ${text}`;
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
    const icon = document.createElement('i');
    icon.className = 'sort-icon ml-1 text-gray-400 ' + (newDir === 'asc' ? 'ri-arrow-up-s-line' : 'ri-arrow-down-s-line');
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
                        r.classList.remove(...SELECTED)
                    );
                    selectedRows = [];
                }
                rows.slice(Math.min(from, to), Math.max(from, to) + 1).forEach(r => {
                    r.classList.add(...SELECTED);
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
                toggleClasses(this, SELECTED.join(' '));
                const idx = selectedRows.indexOf(rowIndex);
                if (idx === -1) selectedRows.push(rowIndex);
                else selectedRows.splice(idx, 1);
                sveselect = 0;
                setSelectAllLabel(false);
            } else {
                table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                    r.classList.remove(...SELECTED)
                );
                toggleClasses(this, SELECTED.join(' '));
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
                    row.classList.add(...SELECTED);
                    if (!selectedRows.includes(rowIndex)) selectedRows.push(rowIndex);
                } else {
                    row.classList.remove(...SELECTED);
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
            !t.closest('#fmContextMenu') &&
            !t.closest('.modal.fade') &&
            !t.closest('.btn-group') &&
            !t.closest('.fmdrawer') &&
            !t.closest('#fmPickerModal') &&
            !wasSelecting
        ) {
            table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                r.classList.remove(...SELECTED)
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
        const tone = id === 'deleteButton' ? 'text-red-600 hover:bg-red-50 dark:text-red-500 dark:hover:bg-red-400/10' : 'text-gray-900 hover:bg-gray-100 dark:text-gray-50 dark:hover:bg-gray-900';
        const sep = id === 'deleteButton' ? '<li role="separator" class="-mx-1 my-1 h-px bg-gray-200 dark:bg-gray-800"></li>' : '';
        return `${sep}<li data-target="${id}" class="flex w-full items-center gap-2 rounded px-2 py-1.5 text-sm cursor-pointer ${tone}"><i class="${icon} text-base leading-none"></i>${label}</li>`;
    }).join('');

    const table = document.getElementById('filemanager_table');

    table.addEventListener('contextmenu', event => {
        const row = event.target.closest('tr.clickable-row');
        if (!row) return;
        event.preventDefault();

        if (!row.classList.contains('selected-row')) {
            table.querySelectorAll('tbody tr.selected-row').forEach(r =>
                r.classList.remove(...SELECTED)
            );
            row.classList.add(...SELECTED);
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
            rows.forEach(r => r.classList.add(...SELECTED));
            selectedRows = Array.from(rows).map((_, i) => String(i));
            sveselect = 1;
            setSelectAllLabel(true);
        } else {
            rows.forEach(r => r.classList.remove(...SELECTED));
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
                .then(data => { cell.textContent = duSizeLabel(data.size); })
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

        const crossed = rows.slice(1).flatMap(r => Array.from(r.querySelectorAll('td:first-child h6, td:first-child h6 a')));
        const crossOut = on => crossed.forEach(el => {
            ['line-through', 'text-red-600', 'dark:text-red-500'].forEach(c => el.classList.toggle(c, on));
        });

        const firstCell = rows[0].querySelector('td');
        const icon = firstCell.querySelector('i')?.outerHTML || '';
        const nameEl = document.createElement('span');
        nameEl.textContent = rows[0].dataset.file;

        inlineEdit(firstCell, `
            <div class="w-4 text-center">${icon}</div>
            <span class="text-sm font-medium text-red-600 line-through dark:text-red-500">${nameEl.innerHTML}</span>
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
                    if (total > 1) saveBtn.innerHTML = `<i class="ri-delete-bin-line"></i> ${done}/${total}`;
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
            saveIcon: 'ri-delete-bin-line',
            danger: true,
            extra: FM.canTrash ? `<label class="inline-flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer"><input type="checkbox" class="fm-inline-skip-trash"> ${FM.i18n.skipTrash}</label>` : '',
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

    // the editor and viewer open in their own centered window like cPanel's, one window per file so reopening focuses it
    function openFileWindow(url, key) {
        const sw = screen.availWidth || 1280, sh = screen.availHeight || 900;
        // a real window on top of the panel, not a near full-screen one
        const w = Math.round(Math.min(1100, Math.max(760, sw * 0.7))), h = Math.round(Math.min(720, Math.max(520, sh * 0.75)));
        const left = Math.max(0, Math.round((sw - w) / 2 + (screen.availLeft || 0)));
        const top = Math.max(0, Math.round((sh - h) / 2 + (screen.availTop || 0)));
        let hash = 0;
        for (const c of key) hash = (hash * 31 + c.charCodeAt(0)) | 0;
        const win = window.open(url, 'opfile' + (hash >>> 0).toString(36), `popup=yes,width=${w},height=${h},left=${left},top=${top}`);
        if (win) win.focus();
        else window.open(url, '_blank'); // popups blocked, fall back to a tab
    }

    // the editor window tells us after a save so sizes and dates in the list are current
    window.addEventListener('message', (e) => {
        if (e.origin === location.origin && e.data && e.data.type === 'op-file-saved') location.reload();
    });

    // VIEW FILE
    document.getElementById('viewButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        let name = row.dataset.file;
        if (name === 'wp-config.php') name = 'wp_temp_openpanel-config.php';
        openFileWindow(`/file-manager/view-file/${encodeURIComponent(FM.pathParam + '/')}${encodeURIComponent(name)}`, 'view:' + FM.pathParam + '/' + name);
    });

    // EDIT FILE
    document.getElementById('editButton').addEventListener('click', () => {
        const row = document.querySelector('#filemanager_table tbody tr.selected-row');
        let name = row.dataset.file;
        if (name === 'wp-config.php') name = 'wp_temp_openpanel-config.php';
        openFileWindow(`/file-manager/edit-file/${encodeURIComponent(FM.pathParam + '/')}${encodeURIComponent(name)}?fullscreen=true`, 'edit:' + FM.pathParam + '/' + name);
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
            <input type="text" aria-label="${FM.i18n.rename}" class="fm-inline-input select-text w-64 block appearance-none rounded-md border border-gray-300 bg-white px-2 py-1 text-xs text-gray-900 shadow-xs outline-hidden transition placeholder:text-gray-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 disabled:border-gray-300 disabled:bg-gray-100 disabled:text-gray-400 aria-[invalid=true]:border-red-500 aria-[invalid=true]:ring-2 aria-[invalid=true]:ring-red-200 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50 dark:placeholder:text-gray-500 dark:focus:border-blue-700 dark:focus:ring-blue-700/30 dark:disabled:border-gray-700 dark:disabled:bg-gray-800 dark:disabled:text-gray-500">
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

    // PERMISSIONS (dialog, one mode for every selected item)
    const openForSelection = () => {
        const rows = Array.from(document.querySelectorAll('#filemanager_table tbody tr.selected-row'));
        if (rows.length) openPermDialog(getSelectedItems(), rows);
    };
    document.getElementById('permButton').addEventListener('click', openForSelection);
    document.getElementById('filemanager_table')?.addEventListener('click', event => {
        if (event.target.closest('[data-perm-open]')) openForSelection();
    });
}

// PERMISSIONS DIALOG

function octalToSymbolic(mode) {
    return mode.split('').map(d => {
        const n = parseInt(d, 8);
        return (n & 4 ? 'r' : '-') + (n & 2 ? 'w' : '-') + (n & 1 ? 'x' : '-');
    }).join('');
}

// a saved mode shows up in the row without a reload: the digits, the rwx tooltip and red when anyone can write
function updatePermCell(row, mode) {
    const sym = row.dataset.permissions.charAt(0) + octalToSymbolic(mode);
    row.dataset.mode = mode;
    row.dataset.permissions = sym;
    const btn = row.querySelector('[data-perm-open]');
    if (!btn) return;
    btn.textContent = mode;
    btn.title = sym;
    const risky = (parseInt(mode[2], 8) & 2) !== 0;
    btn.classList.toggle('text-red-600', risky);
    btn.classList.toggle('dark:text-red-500', risky);
    btn.classList.toggle('text-gray-600', !risky);
    btn.classList.toggle('dark:text-gray-400', !risky);
}

function openPermDialog(items, rows) {
    const modal = document.getElementById('fmPermModal');
    if (!modal) return;
    const input = document.getElementById('fmPermMode');
    const bits = Array.from(modal.querySelectorAll('[data-perm-bit]'));
    const digits = Array.from(modal.querySelectorAll('[data-perm-digit]'));
    const presets = Array.from(modal.querySelectorAll('[data-perm-preset]'));
    const recursive = document.getElementById('fmPermRecursive');
    const recursiveRow = document.getElementById('fmPermRecursiveRow');
    const apply = document.getElementById('fmPermApply');

    const current = rows.map(r => r.dataset.mode || symbolicToOctal(r.dataset.permissions));
    const shared = current.every(v => v === current[0]) ? current[0] : '';
    const hasDirectory = items.some(({ itemType }) => itemType === 'directory');
    const names = items.map(i => i.fileName);
    const errorBox = document.getElementById('fmPermError');
    errorBox.classList.add('hidden');

    document.getElementById('fmPermItems').textContent = names.length === 1
        ? names[0]
        : `${names.length} ${FM.i18n.selected}: ${names.slice(0, 3).join(', ')}${names.length > 3 ? ', …' : ''}`;
    document.getElementById('fmPermWas').classList.toggle('hidden', !shared);
    document.getElementById('fmPermWasMode').textContent = shared;
    const recursiveState = document.getElementById('fmPermRecursiveState');
    recursive.checked = false;
    recursiveState.classList.add('hidden');
    recursiveRow.open = false;
    recursiveRow.classList.toggle('hidden', !hasDirectory);

    // checkboxes, digits, presets and the Apply label all follow one mode
    function show(mode, fromInput) {
        const valid = /^[0-7]{3}$/.test(mode);
        if (!fromInput) input.value = mode;
        input.setAttribute('aria-invalid', valid || mode === '' ? 'false' : 'true');
        if (!valid) {
            apply.disabled = true;
            document.getElementById('fmPermSymbolic').textContent = '';
            document.getElementById('fmPermApplyMode').textContent = '';
            presets.forEach(b => b.setAttribute('aria-pressed', 'false'));
            return;
        }
        apply.disabled = false;
        const flags = octalToSymbolic(mode);
        bits.forEach(b => { b.checked = flags[+b.dataset.permBit] !== '-'; });
        digits.forEach(d => { d.textContent = mode[+d.dataset.permDigit]; });
        document.getElementById('fmPermSymbolic').textContent = flags;
        document.getElementById('fmPermApplyMode').textContent = mode;
        presets.forEach(b => b.setAttribute('aria-pressed', b.dataset.permPreset === mode ? 'true' : 'false'));
    }

    function fromBits() {
        let mode = '';
        for (let i = 0; i < 3; i++) {
            let n = 0;
            if (bits[i * 3].checked) n += 4;
            if (bits[i * 3 + 1].checked) n += 2;
            if (bits[i * 3 + 2].checked) n += 1;
            mode += n;
        }
        show(mode);
    }

    function onInput() {
        // octal only, so drop letters and 8/9 as they're typed
        const clean = input.value.replace(/[^0-7]/g, '').slice(0, 3);
        if (clean !== input.value) input.value = clean;
        show(clean, true);
    }

    function onPreset(e) { show(e.currentTarget.dataset.permPreset); }

    function onApply() {
        const mode = input.value.trim();
        if (!/^[0-7]{3}$/.test(mode)) return;
        if (current.every(v => v === mode) && !recursive.checked) { close(); return; }
        const body = new URLSearchParams([['csrf_token', csrf_token], ['path_param', FM.pathParam], ['permissions', mode]]);
        names.forEach(name => body.append('filename', name));
        if (recursive.checked) body.append('recursive', 'on');
        apply.disabled = true;
        errorBox.classList.add('hidden');
        fetch('/file-manager/permissions', { method: 'POST', headers: { Accept: 'application/json', 'X-CSRF-Token': csrf_token }, body })
            .then(r => r.json())
            .then(res => {
                (res.changed || []).forEach(name => {
                    const row = rows.find(r => r.dataset.file === name);
                    if (row) updatePermCell(row, mode);
                });
                if (res.success) {
                    close();
                    showToast(FM.i18n.permsChanged, 'success');
                    return;
                }
                errorBox.textContent = res.error || FM.i18n.permsFailed;
                errorBox.classList.remove('hidden');
                apply.disabled = false;
            })
            .catch(() => {
                errorBox.textContent = FM.i18n.permsFailed;
                errorBox.classList.remove('hidden');
                apply.disabled = false;
            });
    }

    function onKey(e) {
        if (e.key === 'Escape') close();
        if (e.key === 'Enter' && e.target === input) onApply();
    }

    function onBackdrop(e) { if (e.target === modal) close(); }

    function close() {
        modal.classList.add('hidden');
        recursive.removeEventListener('change', onRecursive);
        bits.forEach(b => b.removeEventListener('change', fromBits));
        presets.forEach(b => b.removeEventListener('click', onPreset));
        input.removeEventListener('input', onInput);
        apply.removeEventListener('click', onApply);
        modal.querySelectorAll('[data-perm-close]').forEach(b => b.removeEventListener('click', close));
        modal.removeEventListener('click', onBackdrop);
        document.removeEventListener('keydown', onKey);
    }

    const onRecursive = () => recursiveState.classList.toggle('hidden', !recursive.checked);
    recursive.addEventListener('change', onRecursive);
    bits.forEach(b => b.addEventListener('change', fromBits));
    presets.forEach(b => b.addEventListener('click', onPreset));
    input.addEventListener('input', onInput);
    apply.addEventListener('click', onApply);
    modal.querySelectorAll('[data-perm-close]').forEach(b => b.addEventListener('click', close));
    modal.addEventListener('click', onBackdrop);
    document.addEventListener('keydown', onKey);

    show(shared || current[0] || '644');
    modal.classList.remove('hidden');
    input.focus();
    input.select();
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
        nameInput.removeAttribute('aria-invalid');
    }

    function setDest(path) {
        cur = path;
        destInput.value = '/' + cur;
        destInput.removeAttribute('aria-invalid');
        // moving into the folder the items already live in is a no-op
        confirmBtn.disabled = (action === 'move' || action === 'copy') && (cur === FM.pathParam.replace(/^\/|\/$/g, '') || isBlocked(cur));
    }

    function renderCrumbs() {
        const parts = cur ? cur.split('/') : [];
        const crumbs = [`<button type="button" data-path="" aria-label="/" class="text-gray-500 hover:text-gray-900 dark:hover:text-gray-50 cursor-pointer"><i class="ri-home-4-line"></i></button>`];
        parts.forEach((part, i) => {
            crumbs.push(`<span class="text-gray-400 dark:text-gray-600">/</span><button type="button" data-path="${escHTML(parts.slice(0, i + 1).join('/'))}" class="text-gray-900 dark:text-gray-50 hover:underline hover:underline-offset-4 cursor-pointer truncate">${escHTML(part)}</button>`);
        });
        $('fmPickerCrumbs').innerHTML = crumbs.join('');
    }

    async function load(path) {
        list.innerHTML = `<li class="px-3 py-2 text-gray-500">${FM.i18n.working}</li>`;
        let folders;
        try {
            const res = await fetch('/json/folders/' + path.split('/').map(encodeURIComponent).join('/'));
            if (!res.ok) throw new Error();
            folders = (await res.json()).folders || [];
            // the malware scanner's quarantine is not a place to put files
            if (!path) folders = folders.filter(f => f.name !== '.quarantine');
        } catch (e) {
            destInput.setAttribute('aria-invalid', 'true');
            showToast(FM.i18n.folderNotFound, 'error');
            if (path !== cur) return load(cur);
            folders = [];
        }

        setDest(path);
        renderCrumbs();
        const rows = [];
        if (cur) {
            rows.push(`<li><button type="button" data-path="${escHTML(cur.split('/').slice(0, -1).join('/'))}" class="flex w-full items-center gap-2 px-3 py-2 text-left text-gray-500 hover:bg-gray-50 dark:hover:bg-gray-900/60 cursor-pointer"><i class="ri-corner-left-up-line"></i> ..</button></li>`);
        }
        folders.sort((x, y) => x.name.localeCompare(y.name)).forEach(f => {
            const p = joinPath(cur, f.name);
            const off = isBlocked(p);
            rows.push(`<li><button type="button" data-path="${escHTML(p)}" ${off ? 'disabled' : ''} class="flex w-full items-center gap-2 px-3 py-2 text-left ${off ? 'cursor-not-allowed text-gray-400 dark:text-gray-600' : 'hover:bg-gray-50 dark:hover:bg-gray-900/60 cursor-pointer'}"><i class="ri-folder-fill text-orange-500"></i><span class="truncate min-w-0">${escHTML(f.name)}</span>${f.has_subfolders && !off ? '<i class="ri-arrow-right-s-line ml-auto text-gray-400"></i>' : ''}</button></li>`);
        });
        if (!folders.length) rows.push(`<li class="px-3 py-2 text-gray-500">${FM.i18n.noSubfolders}</li>`);
        list.innerHTML = rows.join('');
    }

    function showStatus(kind, html) {
        status.className = 'rounded-md p-3 text-sm ' + (kind === 'error' ? 'bg-red-50 text-red-900 dark:bg-red-950/70 dark:text-red-500' : 'bg-gray-100 text-gray-900 dark:bg-gray-800/70 dark:text-gray-400');
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
        li.innerHTML = `<i class="ri-folder-fill text-orange-500"></i><input type="text" data-no-live-validate aria-label="${FM.i18n.newFolder}" class="select-text w-full block appearance-none rounded-md border border-gray-300 bg-white px-2 py-1 text-xs text-gray-900 shadow-xs outline-hidden transition placeholder:text-gray-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 disabled:border-gray-300 disabled:bg-gray-100 disabled:text-gray-400 aria-[invalid=true]:border-red-500 aria-[invalid=true]:ring-2 aria-[invalid=true]:ring-red-200 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50 dark:placeholder:text-gray-500 dark:focus:border-blue-700 dark:focus:ring-blue-700/30 dark:disabled:border-gray-700 dark:disabled:bg-gray-800 dark:disabled:text-gray-500">`;
        list.prepend(li);
        const input = li.querySelector('input');
        input.focus();
        input.addEventListener('keydown', async e => {
            e.stopPropagation();
            if (e.key === 'Escape') return li.remove();
            if (e.key !== 'Enter') return;
            const name = input.value.trim();
            if (!name || name.includes('/')) { input.setAttribute('aria-invalid', 'true'); return; }
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
            nameInput.setAttribute('aria-invalid', 'true');
            return nameInput.focus();
        }
        runAction();
    };
    nameInput.oninput = () => { nameInput.removeAttribute('aria-invalid'); };
    nameInput.onkeydown = e => { if (e.key === 'Enter') { e.preventDefault(); confirmBtn.click(); } };
    modal.querySelectorAll('[data-picker-close]').forEach(b => { b.onclick = close; });
    modal.onclick = e => { if (e.target === modal) close(); };
    document.addEventListener('keydown', onKey);

    modal.classList.remove('hidden');
    load(cur);
    if (action === 'compress') { nameInput.focus(); nameInput.select(); } else confirmBtn.focus();
}

// UTILITIES

// small recipe buttons for the inline editors
const FM_BTN_BASE = 'relative inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-md border px-2.5 py-1.5 text-center text-xs font-medium shadow-xs transition-all duration-100 ease-in-out disabled:pointer-events-none disabled:shadow-none outline outline-0 outline-offset-2 outline-blue-500 focus-visible:outline-2';
const FM_BTN_PRIMARY = FM_BTN_BASE + ' border-transparent bg-blue-500 text-white hover:bg-blue-600 disabled:bg-blue-300 dark:disabled:bg-blue-800 dark:disabled:text-blue-400';
const FM_BTN_DANGER = FM_BTN_BASE + ' border-transparent bg-red-600 text-white hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600';
const FM_BTN_GHOST = FM_BTN_BASE + ' border-transparent bg-transparent shadow-none text-gray-900 hover:bg-gray-100 dark:text-gray-50 dark:hover:bg-gray-800/80';
const FM_BTN_SECONDARY = FM_BTN_BASE + ' border-gray-300 bg-white text-gray-900 hover:bg-gray-50 disabled:text-gray-400 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50 dark:hover:bg-gray-900/60 dark:disabled:text-gray-600';


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
    // whole bytes stay whole, bigger units get one decimal: 62 B, 4.5 KB
    return (i === 0 ? Math.round(bytes) : bytes.toFixed(1)) + ' ' + units[i];
}

// du -h prints 4.0K / 12M / 0, the table shows 4.0 KB / 12 MB / 0 B like the file sizes
function duSizeLabel(size) {
    const m = String(size).trim().match(/^([\d.]+)\s*([KMGTP])?i?B?$/i);
    if (!m) return size;
    return m[1] + ' ' + (m[2] ? m[2].toUpperCase() + 'B' : 'B');
}

// swaps a table cell for an input + save/cancel, onSave returns false to flag the input red, true to just close
let closeInlineEdit = null;

function inlineEdit(cell, innerHTML, onSave, opts = {}) {
    if (closeInlineEdit) closeInlineEdit();

    const buttons = `
            ${opts.extra || ''}
            <button type="button" data-act="save" class="${opts.danger ? FM_BTN_DANGER : FM_BTN_PRIMARY}"><i class="${opts.saveIcon || 'ri-check-line'}"></i> ${opts.saveLabel || FM.i18n.save}</button>
            <button type="button" data-act="cancel" class="${FM_BTN_GHOST}">${FM.i18n.cancel}</button>`;

    const editor = document.createElement('td');
    editor.className = cell.className.replace(/\b(text-center|font-mono|text-xs)\b/g, '').replace(/\bpy-2\.5\b/, 'py-1.5') + ' fm-inline-editor';
    // last column has no room to the right, so float the buttons left over the neighbouring cells
    editor.innerHTML = opts.buttonsBefore ? `
        <div class="relative flex items-center justify-center py-0.5">
            <div class="absolute top-1/2 right-full z-10 mr-2 flex -translate-y-1/2 items-center gap-2 whitespace-nowrap rounded-md bg-white py-1 pl-2 dark:bg-gray-950">${buttons}</div>
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
            input.setAttribute('aria-invalid', 'true');
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
        input.addEventListener('input', () => { input.removeAttribute('aria-invalid'); });
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

function symbolicToOctal(symbolic) {
    if (!symbolic || symbolic.length < 9) return '';
    // setuid/setgid/sticky (s, t) still mean execute is on, capital S/T mean it's off
    const map = { r: 4, w: 2, x: 1, s: 1, t: 1, S: 0, T: 0, '-': 0 };
    return symbolic.slice(-9).match(/.{3}/g)
        .map(p => p.split('').reduce((s, c) => s + (map[c] || 0), 0))
        .join('');
}

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
