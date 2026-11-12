document.addEventListener('DOMContentLoaded', function () {
    fetch('/domains/capitalize')
        .then(response => response.json())
        .then(data => {
            if (!data.error) {
                const excludedElements = [
                    document.getElementById('filemanager_table'),
                    document.getElementById('editor-container'),
                    document.getElementById('editor'),
                    document.getElementById('zone_content'),
                    ...Array.from(document.getElementsByClassName('domain_link'))
                ].filter(Boolean); // Remove any nulls

                // TreeWalker to go through all text nodes
                const walker = document.createTreeWalker(
                    document.body,
                    NodeFilter.SHOW_TEXT,
                    {
                        acceptNode: function (node) {
                            for (const excluded of excludedElements) {
                                if (excluded.contains(node)) {
                                    return NodeFilter.FILTER_REJECT;
                                }
                            }
                            return NodeFilter.FILTER_ACCEPT;
                        }
                    }
                );

                let node;
                while ((node = walker.nextNode())) {
                    let replaced = node.nodeValue;
                    for (const [domain, capitalized] of Object.entries(data)) {
                        const regex = new RegExp(domain, 'g');
                        replaced = replaced.replace(regex, capitalized);
                    }
                    if (replaced !== node.nodeValue) {
                        node.nodeValue = replaced;
                    }
                }
            }
        })
        .catch(error => console.error('Error fetching capitalized domains:', error));
});
