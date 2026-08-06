// Load Clipboard.js from CDN
var script = document.createElement("script");
script.src = "https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.10/clipboard.min.js";
script.onload = function () {
    var breadcrumbContainer = document.querySelector("nav[aria-label='FileManager Breadcrumbs'] ol");
    if (!breadcrumbContainer) return;

    var breadcrumbItems = breadcrumbContainer.querySelectorAll("li");
    var maxVisibleItems = 5;
    var chevronIcons = breadcrumbContainer.querySelectorAll("svg");

    // Create "Copy Path to Clipboard" button
    var copyButton = document.createElement("li");
    copyButton.innerHTML = '<a href="#" class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300" id="copy-button" title="Copy Path to Clipboard" data-clipboard-text=""><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-copy" viewBox="0 0 16 16"> <path fill-rule="evenodd" d="M4 2a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2zm2-1a1 1 0 0 0-1 1v8a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V2a1 1 0 0 0-1-1zM2 5a1 1 0 0 0-1 1v8a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1v-1h1v1a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h1v1z"/> </svg></a>';
    breadcrumbContainer.appendChild(copyButton);

    // Hide middle breadcrumb items if they exceed the limit
    if (breadcrumbItems.length > maxVisibleItems) {
        for (var i = 1; i < breadcrumbItems.length - 1; i++) {
            breadcrumbItems[i].style.display = "none";
        }

        // Hide middle chevron icons but keep the first and last
        chevronIcons.forEach((icon, index) => {
            if (index > 0 && index < chevronIcons.length - 1) {
                icon.style.display = "none";
            }
        });

        // Create ellipsis link
        var ellipsis = document.createElement("li");
        ellipsis.innerHTML = '<a href="#" class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300" id="breadcrumb-ellipsis" title="Click to display the full path"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-three-dots" viewBox="0 0 16 16"> <path d="M3 9.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3m5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3m5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3"/> </svg></a>';
        breadcrumbContainer.insertBefore(ellipsis, breadcrumbItems[1]);

        ellipsis.addEventListener("click", function (event) {
            event.preventDefault();
            breadcrumbItems.forEach(item => item.style.display = "");
            chevronIcons.forEach(icon => icon.style.display = "");
            ellipsis.style.display = "none";
        });
    }

    var clipboard = new ClipboardJS("#copy-button", {
        text: function () {
            var fullPath = "";
            breadcrumbItems.forEach((item, index) => {
                var trimmedText = item.textContent.trim();
                if (index > 1) {
                    fullPath += '/';
                }
                fullPath += trimmedText;
            });
            return fullPath;
        }
    });

    clipboard.on("success", function () {
        const toastMessage = "Path copied to clipboard!";
        showToast(toastMessage, 'success');
    });
};

document.head.appendChild(script);
