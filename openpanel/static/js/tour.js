/* OpenPanel UI product tour - cross-page step-by-step walkthrough built on driver.js */
(function () {
    var STORAGE_ACTIVE = 'op_tour_active';
    var STORAGE_INDEX = 'op_tour_index';

    function normalizePath(p) {
        if (p === '/') return '/dashboard';
        return p.replace(/\/+$/, '') || '/dashboard';
    }

    function currentPath() {
        return normalizePath(window.location.pathname);
    }

    function getStoredIndex() {
        var raw = localStorage.getItem(STORAGE_INDEX);
        return raw === null ? 0 : parseInt(raw, 10);
    }

    function setState(active, index) {
        if (active) {
            localStorage.setItem(STORAGE_ACTIVE, '1');
            localStorage.setItem(STORAGE_INDEX, String(index));
        } else {
            localStorage.removeItem(STORAGE_ACTIVE);
            localStorage.removeItem(STORAGE_INDEX);
        }
    }

    function isActive() {
        return localStorage.getItem(STORAGE_ACTIVE) === '1';
    }

    function markComplete() {
        fetch('/dashboard/tour/complete', {
            method: 'POST',
            headers: { 'X-CSRF-TOKEN': window.csrf_token || '' }
        }).catch(function () {});
    }

    function runHook(name) {
        if (name && window.TOUR_HOOKS && typeof window.TOUR_HOOKS[name] === 'function') {
            window.TOUR_HOOKS[name]();
        }
    }

    var driverObj = null;
    var currentPageStartIndex = 0;
    var currentPageEndIndex = 0;

    function findPageStart(globalIndex) {
        var steps = window.TOUR_STEPS || [];
        var path = normalizePath(steps[globalIndex].path);
        var j = globalIndex;
        while (j > 0 && normalizePath(steps[j - 1].path) === path) j--;
        return j;
    }

    function goToAdjacentPage(direction) {
        var steps = window.TOUR_STEPS || [];
        if (!steps.length) return;
        if (direction > 0) {
            if (currentPageEndIndex >= steps.length) return;
            setState(true, currentPageEndIndex);
            window.location.href = normalizePath(steps[currentPageEndIndex].path);
        } else {
            if (currentPageStartIndex <= 0) return;
            var target = findPageStart(currentPageStartIndex - 1);
            setState(true, target);
            window.location.href = normalizePath(steps[target].path);
        }
    }

    document.addEventListener('keyup', function (e) {
        if (!driverObj) return;
        if (e.key === 'ArrowDown') goToAdjacentPage(1);
        else if (e.key === 'ArrowUp') goToAdjacentPage(-1);
    });

    function addKeyboardHint(popover) {
        if (!popover || !popover.footer || popover.footer.querySelector('.op-tour-hint')) return;
        var hint = document.createElement('div');
        hint.className = 'op-tour-hint';
        hint.style.cssText = 'width:100%;font-size:11px;color:#9ca3af;margin-top:6px;text-align:center;';
        hint.textContent = '← → steps · ↑ ↓ pages';
        popover.footer.appendChild(hint);
    }

    function buildAndShow(startIndex) {
        var steps = window.TOUR_STEPS || [];
        var path = currentPath();
        var pageSteps = [];
        var i = startIndex;
        while (i < steps.length && normalizePath(steps[i].path) === path) {
            pageSteps.push(steps[i]);
            i++;
        }
        var nextIndexAfterPage = i;
        currentPageStartIndex = startIndex;
        currentPageEndIndex = nextIndexAfterPage;

        var visibleSteps = pageSteps.filter(function (s) {
            return document.querySelector(s.element);
        });
        if (!visibleSteps.length) {
            setState(false);
            return;
        }

        var driverSteps = visibleSteps.map(function (step, localIdx) {
            var isLastOnPage = localIdx === visibleSteps.length - 1;
            var nextVisible = visibleSteps[localIdx + 1];
            var isFinalStep = isLastOnPage && nextIndexAfterPage >= steps.length;
            var crossesToNextPage = isLastOnPage && !isFinalStep;

            var popover = {
                title: step.title,
                description: step.description,
                side: step.side || 'bottom',
                align: step.align || 'start',
                showButtons: ['next', 'close'],
                nextBtnText: isFinalStep ? 'Finish' : 'Next'
                // onNextClick/onCloseClick must be omitted entirely (not set to undefined)
                // when not overridden - driver.js spreads the original popover config over
                // its own computed defaults, so an explicit `undefined` here would wipe out
                // its built-in "advance"/"close" handlers and the buttons would do nothing.
                // Leaving them unset lets driver.js fall back to its defaults, which already
                // call destroy() on Finish/close/Escape/overlay-click - all of which trigger
                // the top-level onDestroyed cleanup below.
            };

            if (crossesToNextPage) {
                // moving to a different page: persist progress and navigate, without
                // destroying the current driver instance (the page unload does that for us)
                popover.onNextClick = function () {
                    setState(true, nextIndexAfterPage);
                    window.location.href = normalizePath(steps[nextIndexAfterPage].path);
                };
            } else if (nextVisible && nextVisible.beforeShow) {
                popover.onNextClick = function () {
                    runHook(nextVisible.beforeShow);
                    setTimeout(function () { driverObj.moveNext(); }, 60);
                };
            }

            // _tourStep keeps a reference back to our own step config (for beforeHide) -
            // driver.js only reads `element`/`popover` off this object, so this extra key
            // is harmless to it.
            return { element: step.element, popover: popover, _tourStep: step };
        });

        driverObj = window.driver.js.driver({
            allowClose: true,
            allowKeyboardControl: true,
            overlayOpacity: 0.6,
            stagePadding: 4,
            smoothScroll: true,
            steps: driverSteps,
            onPopoverRender: addKeyboardHint,
            // unlike onDeselected (which fires on every step-to-step transition, since
            // driver.js uses it to deselect the outgoing element before highlighting the
            // next one), onDestroyed only fires when the tour is actually torn down: Finish,
            // the X button, Escape, or clicking the overlay. That makes it the one reliable
            // place to clear our state - clearing it on every transition would make a page
            // refresh mid-tour forget where the user was.
            onDestroyed: function (activeElement, activeStep) {
                setState(false);
                markComplete();
                runHook(activeStep && activeStep._tourStep && activeStep._tourStep.beforeHide);
            }
        });

        setState(true, startIndex);
        if (visibleSteps[0].beforeShow) runHook(visibleSteps[0].beforeShow);
        driverObj.drive();
    }

    function start(fromCurrentPage) {
        var steps = window.TOUR_STEPS || [];
        if (!steps.length) return;
        var path = currentPath();
        var idx = 0;
        if (fromCurrentPage) {
            var found = steps.findIndex(function (s) { return normalizePath(s.path) === path; });
            idx = found === -1 ? 0 : found;
        }
        var targetPath = normalizePath(steps[idx].path);
        if (targetPath !== path) {
            setState(true, idx);
            window.location.href = targetPath;
            return;
        }
        buildAndShow(idx);
    }

    document.addEventListener('DOMContentLoaded', function () {
        if (isActive()) buildAndShow(getStoredIndex());
    });

    window.OPTour = { start: start, markComplete: markComplete };
})();
