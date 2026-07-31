// Tailwind's content glob (tailwind.config.js) only scans ./src and ./docs,
// so these classes live here (not inline in docusaurus.config.js) to make
// sure Tailwind actually generates the CSS for them.
module.exports = {
    button: "!bg-gray-100 dark:!bg-gray-800 !border !border-gray-300 dark:!border-gray-700 !rounded-lg !text-gray-600 dark:!text-gray-300 hover:!bg-gray-200 dark:hover:!bg-gray-700 !shadow-none !font-normal",
    dropdown:
        "!bg-gray-0 dark:!bg-gray-900 !border !border-gray-300 dark:!border-gray-700 !rounded-lg !shadow-lg",
    dropdownItem:
        "!border-gray-200 dark:!border-gray-700 hover:!bg-gray-100 dark:hover:!bg-gray-700",
};
