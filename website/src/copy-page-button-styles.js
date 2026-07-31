// Tailwind's content glob (tailwind.config.js) only scans ./src and ./docs,
// so these classes live here (not passed inline) to make sure Tailwind
// actually generates the CSS for them.
//
// Matches src/refine-theme/common-github-star-button.tsx ("Get Support")
// exactly, since the copy-page button sits in its old header slot now.
module.exports = {
    button: "!bg-transparent !text-sm !text-gray-500 dark:!text-gray-400 !rounded-[32px] !border !border-solid !border-gray-300 dark:!border-gray-700 !gap-2 !py-2 !pl-2.5 !pr-4 !shadow-none !font-normal",
    dropdown:
        "!bg-gray-0 dark:!bg-gray-900 !border !border-gray-300 dark:!border-gray-700 !rounded-lg !shadow-lg",
    dropdownItem:
        "!border-gray-200 dark:!border-gray-700 hover:!bg-gray-100 dark:hover:!bg-gray-700",
};
