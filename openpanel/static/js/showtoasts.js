/*
TOASTS EXAMPLES:

// default
showToast(toastMessage, 'success');

// add report button with a link
showToast(toastMessage, 'error', true, 'https://openpanel.com/report?page=xxx&error=yyyy&ver=zzz');

// do not dismiss toast after 5sec
showToast(toastMessage, 'warning', true);

// loading toast that is dismissed when another toast pops
showToast(toastMessage, 'loading', true);

*/

// needed to dismiss progress toasts on another
let activeToasts = [];
let toastMessage;

function removeToast(toastElement) {
    // Check if toast is already removed to avoid duplicate removal
    if (!toastElement || !toastElement.parentElement) {
        return;
    }

    // Find and remove the toast from the activeToasts array
    activeToasts = activeToasts.filter((toast) => {
        if (toast.element === toastElement) {
            toastElement.remove(); // Remove from DOM
            return false; // Remove from activeToasts array
        }
        return true; // Keep other toasts
    });
}


// js toasts displayed by backend notifications and on page ajax responses
function showToast(message, type, isPermanent = false, link = false) {
    const toastContainer = document.getElementById('toast-container');

    if (!toastContainer) {
        console.error('Toast container not found!');
        return;
    }


    // {label, href} or {label, onClick} adds a neutral action like Open or Undo, a plain string keeps the old red Cancel link
    let action = null;
    if (link && typeof link === 'object') {
        action = link;
        link = `
            ${action.href
                ? `<a href="${action.href}" data-toast-action class="flex flex-1 items-center justify-center px-6 text-sm font-semibold transition-colors hover:bg-gray-50 hover:dark:bg-gray-900/30 active:bg-gray-100 active:dark:bg-gray-800 text-blue-600 dark:text-blue-400"></a>`
                : `<button type="button" data-toast-action class="flex flex-1 items-center justify-center px-6 text-sm font-semibold transition-colors hover:bg-gray-50 hover:dark:bg-gray-900/30 active:bg-gray-100 active:dark:bg-gray-800 text-blue-600 dark:text-blue-400"></button>`}
            <div class="h-px w-full bg-gray-200 dark:bg-gray-800"></div>
        `;
    } else if (link) {
        link = `
            <a href="${link}" class="flex flex-1 items-center justify-center px-6 text-sm font-semibold transition-colors hover:bg-gray-50 hover:dark:bg-gray-900/30 active:bg-gray-100 active:dark:bg-gray-800 text-red-600 dark:text-red-500">${window.ToastI18n?.cancel || 'Cancel'}</a>
            <div class="h-px w-full bg-gray-200 dark:bg-gray-800"></div>
        `;
    } else {
        link = `
        
        `;
    }


    // Remove any existing toasts when loader..
    activeToasts = activeToasts.filter((toast) => {
        if (toast.type === 'loading' || toast.type === 'uploading') {
            toast.element.remove(); // Remove the loading toast element
            return false; // Remove from activeToasts array
        }
        return true; // Keep other toasts
    });
    

    // Define different icon and background classes based on the toast type
    const typeConfig = {
        success: {
            icon: `
                <i class="ri-checkbox-circle-fill text-xl leading-none" aria-hidden="true"></i>
            `,
            bgColor: 'text-emerald-500 dark:text-emerald-500',
            text: window.ToastI18n?.success || 'Success',
        },
        error: {
            icon: `
                <i class="ri-error-warning-fill text-xl leading-none" aria-hidden="true"></i>
            `,
            bgColor: 'text-red-500 dark:text-red-500',
            text: window.ToastI18n?.error || 'Error',
        },
        info: {
            icon: `
                <i class="ri-information-fill text-xl leading-none" aria-hidden="true"></i>
            `,
            bgColor: 'text-blue-500 dark:text-blue-500',
            text: window.ToastI18n?.info || 'Info',
        },
        uploading: {
            icon: `
                <i class="ri-loader-4-line text-xl leading-none" aria-hidden="true"></i>
            `,
            bgColor: 'animate-spin text-grey-600 dark:text-grey-600',
            text: window.ToastI18n?.uploading || 'Uploading',
        },
        loading: {
            icon: `
                <i class="ri-loader-4-line text-xl leading-none" aria-hidden="true"></i>
            `,
            bgColor: 'animate-spin text-grey-600 dark:text-grey-600',
            text: window.ToastI18n?.loading || 'Loading',
        },
        warning: {
            icon: `
                <i class="ri-alert-fill text-xl leading-none" aria-hidden="true"></i>
            `,
            bgColor: 'text-yellow-500 dark:text-yellow-500',
            text: window.ToastI18n?.warning || 'Warning',
        },
    };

    const { icon, bgColor, text } = typeConfig[type] || typeConfig['info'];

    const toastElement = document.createElement('div');
    toastElement.className = `flex h-fit min-h-16 overflow-hidden rounded-md border shadow-lg shadow-black/5 bg-white dark:bg-[#090E1A] border-gray-200 dark:border-gray-800 data-[swipe=cancel]:translate-x-0 data-[swipe=end]:translate-x-[var(--radix-toast-swipe-end-x)] data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)] data-[swipe=move]:transition-none data-[state=open]:animate-slideLeftAndFade data-[state=closed]:animate-hide mt-0`;
    toastElement.setAttribute('role', 'alert');


    toastElement.innerHTML = `
        <div class="flex flex-1 items-start gap-3 p-4 border-r border-gray-200 dark:border-gray-800">
            <svg class="w-6 h-6 ${bgColor} rounded-full p-1" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" fill="currentColor" aria-hidden="true">
                ${icon}
            </svg>
            <div class="flex flex-col gap-1">
                <div class="text-sm font-semibold text-gray-900 dark:text-gray-50">${text}</div>
                <div class="text-sm text-gray-600 dark:text-gray-400" id="toastMessage">${message}</div>
            </div>
        </div>
        <div class="flex flex-col">
            ${link}
            <button type="button" class="flex flex-1 items-center justify-center px-6 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-900/30 active:bg-gray-100 h-full" aria-label="Close">
                ${window.ToastI18n?.close || 'Close'}
            </button>
        </div>
    `;

    if (action) {
        const actionEl = toastElement.querySelector('[data-toast-action]');
        actionEl.textContent = action.label;
        if (action.onClick) {
            actionEl.addEventListener('click', () => {
                removeToast(toastElement);
                action.onClick();
            });
        }
    }

    // Append the toast to the container
    toastContainer.appendChild(toastElement);

    // add to active toasts
    activeToasts.push({ type, element: toastElement });

     // Automatically remove the toast after 5 seconds if not permanent
    if (!isPermanent) {
        const autoRemove = setTimeout(() => {
            removeToast(toastElement);
        }, 5000);

        // Add click event to the close button to manually dismiss the toast
        toastElement.querySelector('button[aria-label="Close"]').addEventListener('click', () => {
            clearTimeout(autoRemove); // TODO
            removeToast(toastElement); 
        });
    } else {
        // Add click event to the close button for permanent toast
        toastElement.querySelector('button[aria-label="Close"]').addEventListener('click', () => {
            removeToast(toastElement);
        });
    }
}
