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


    if (link) {
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
                <svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 8.207-4 4a1 1 0 0 1-1.414 0l-2-2a1 1 0 0 1 1.414-1.414L9 10.586l3.293-3.293a1 1 0 0 1 1.414 1.414Z"/>
                </svg>
            `,
            bgColor: 'text-emerald-500 dark:text-emerald-500',
            text: window.ToastI18n?.success || 'Success',
        },
        error: {
            icon: `
                <svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 11.793a1 1 0 1 1-1.414 1.414L10 11.414l-2.293 2.293a1 1 0 0 1-1.414-1.414L8.586 10 6.293 7.707a1 1 0 0 1 1.414-1.414L10 8.586l2.293-2.293a1 1 0 0 1 1.414 1.414L11.414 10l2.293 2.293Z"/>
                </svg>
            `,
            bgColor: 'text-red-500 dark:text-red-500',
            text: window.ToastI18n?.error || 'Error',
        },
        info: {
            icon: `
                <svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12C22 17.5228 17.5228 22 12 22ZM11 11V17H13V11H11ZM11 7V9H13V7H11Z"></path>
                </svg>
            `,
            bgColor: 'text-blue-500 dark:text-blue-500',
            text: window.ToastI18n?.info || 'Info',
        },
        uploading: {
            icon: `
                <svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M12 2C12.5523 2 13 2.44772 13 3V6C13 6.55228 12.5523 7 12 7C11.4477 7 11 6.55228 11 6V3C11 2.44772 11.4477 2 12 2ZM12 17C12.5523 17 13 17.4477 13 18V21C13 21.5523 12.5523 22 12 22C11.4477 22 11 21.5523 11 21V18C11 17.4477 11.4477 17 12 17ZM22 12C22 12.5523 21.5523 13 21 13H18C17.4477 13 17 12.5523 17 12C17 11.4477 17.4477 11 18 11H21C21.5523 11 22 11.4477 22 12ZM7 12C7 12.5523 6.55228 13 6 13H3C2.44772 13 2 12.5523 2 12C2 11.4477 2.44772 11 3 11H6C6.55228 11 7 11.4477 7 12ZM19.0711 19.0711C18.6805 19.4616 18.0474 19.4616 17.6569 19.0711L15.5355 16.9497C15.145 16.5592 15.145 15.9261 15.5355 15.5355C15.9261 15.145 16.5592 15.145 16.9497 15.5355L19.0711 17.6569C19.4616 18.0474 19.4616 18.6805 19.0711 19.0711ZM8.46447 8.46447C8.07394 8.85499 7.44078 8.85499 7.05025 8.46447L4.92893 6.34315C4.53841 5.95262 4.53841 5.31946 4.92893 4.92893C5.31946 4.53841 5.95262 4.53841 6.34315 4.92893L8.46447 7.05025C8.85499 7.44078 8.85499 8.07394 8.46447 8.46447ZM4.92893 19.0711C4.53841 18.6805 4.53841 18.0474 4.92893 17.6569L7.05025 15.5355C7.44078 15.145 8.07394 15.145 8.46447 15.5355C8.85499 15.9261 8.85499 16.5592 8.46447 16.9497L6.34315 19.0711C5.95262 19.4616 5.31946 19.4616 4.92893 19.0711ZM15.5355 8.46447C15.145 8.07394 15.145 7.44078 15.5355 7.05025L17.6569 4.92893C18.0474 4.53841 18.6805 4.53841 19.0711 4.92893C19.4616 5.31946 19.4616 5.95262 19.0711 6.34315L16.9497 8.46447C16.5592 8.85499 15.9261 8.85499 15.5355 8.46447Z"></path>
                </svg>
            `,
            bgColor: 'animate-spin text-grey-600 dark:text-grey-600',
            text: window.ToastI18n?.uploading || 'Uploading',
        },
        loading: {
            icon: `
                <svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M12 2C12.5523 2 13 2.44772 13 3V6C13 6.55228 12.5523 7 12 7C11.4477 7 11 6.55228 11 6V3C11 2.44772 11.4477 2 12 2ZM12 17C12.5523 17 13 17.4477 13 18V21C13 21.5523 12.5523 22 12 22C11.4477 22 11 21.5523 11 21V18C11 17.4477 11.4477 17 12 17ZM22 12C22 12.5523 21.5523 13 21 13H18C17.4477 13 17 12.5523 17 12C17 11.4477 17.4477 11 18 11H21C21.5523 11 22 11.4477 22 12ZM7 12C7 12.5523 6.55228 13 6 13H3C2.44772 13 2 12.5523 2 12C2 11.4477 2.44772 11 3 11H6C6.55228 11 7 11.4477 7 12ZM19.0711 19.0711C18.6805 19.4616 18.0474 19.4616 17.6569 19.0711L15.5355 16.9497C15.145 16.5592 15.145 15.9261 15.5355 15.5355C15.9261 15.145 16.5592 15.145 16.9497 15.5355L19.0711 17.6569C19.4616 18.0474 19.4616 18.6805 19.0711 19.0711ZM8.46447 8.46447C8.07394 8.85499 7.44078 8.85499 7.05025 8.46447L4.92893 6.34315C4.53841 5.95262 4.53841 5.31946 4.92893 4.92893C5.31946 4.53841 5.95262 4.53841 6.34315 4.92893L8.46447 7.05025C8.85499 7.44078 8.85499 8.07394 8.46447 8.46447ZM4.92893 19.0711C4.53841 18.6805 4.53841 18.0474 4.92893 17.6569L7.05025 15.5355C7.44078 15.145 8.07394 15.145 8.46447 15.5355C8.85499 15.9261 8.85499 16.5592 8.46447 16.9497L6.34315 19.0711C5.95262 19.4616 5.31946 19.4616 4.92893 19.0711ZM15.5355 8.46447C15.145 8.07394 15.145 7.44078 15.5355 7.05025L17.6569 4.92893C18.0474 4.53841 18.6805 4.53841 19.0711 4.92893C19.4616 5.31946 19.4616 5.95262 19.0711 6.34315L16.9497 8.46447C16.5592 8.85499 15.9261 8.85499 15.5355 8.46447Z"></path>
                </svg>
            `,
            bgColor: 'animate-spin text-grey-600 dark:text-grey-600',
            text: window.ToastI18n?.loading || 'Loading',
        },
        warning: {
            icon: `
                <svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM10 15a1 1 0 1 1 0-2 1 1 0 0 1 0 2Zm1-4a1 1 0 0 1-2 0V6a1 1 0 0 1 2 0v5Z"/>
                </svg>
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
        toastElement.querySelector('button').addEventListener('click', () => {
            clearTimeout(autoRemove); // TODO
            removeToast(toastElement); 
        });
    } else {
        // Add click event to the close button for permanent toast
        toastElement.querySelector('button').addEventListener('click', () => {
            removeToast(toastElement);
        });
    }
}
