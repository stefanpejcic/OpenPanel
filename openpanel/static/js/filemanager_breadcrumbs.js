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
    copyButton.className = "flex";
    copyButton.innerHTML = '<a href="#" class="inline-flex size-6 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800/80 dark:hover:text-gray-50" id="copy-button" title="Copy path to clipboard" aria-label="Copy path to clipboard" data-clipboard-text=""><i class="ri-file-copy-line"></i></a>';
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
        ellipsis.className = "flex";
        ellipsis.innerHTML = '<a href="#" class="inline-flex size-6 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800/80 dark:hover:text-gray-50" id="breadcrumb-ellipsis" title="Click to display the full path" aria-label="Click to display the full path"><i class="ri-more-line"></i></a>';
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
        const toastMessage = (typeof FM !== "undefined" && FM.i18n.pathCopied) || "Path copied to clipboard!";
        showToast(toastMessage, 'success');
    });
};

document.head.appendChild(script);
