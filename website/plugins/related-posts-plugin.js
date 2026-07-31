const fs = require("fs");
const path = require("path");
const {
    parseMarkdownFile,
    DEFAULT_PARSE_FRONT_MATTER,
    normalizeUrl,
} = require("@docusaurus/utils");

// Same convention @docusaurus/plugin-content-blog uses to derive a post's
// date from its filename when frontmatter has no explicit `date`.
const DATE_FILENAME_REGEX = /^(\d{4})-(\d{1,2})-(\d{1,2})-.*\.mdx?$/;

function getMultipleRandomElements(arr, num) {
    const shuffled = [...arr].sort(() => 0.5 - Math.random());
    return shuffled.slice(0, num);
}

function loadPost(blogDir, routeBasePath, file) {
    const filePath = path.join(blogDir, file);
    const fileContent = fs.readFileSync(filePath, "utf8");

    return parseMarkdownFile({
        filePath,
        fileContent,
        parseFrontMatter: DEFAULT_PARSE_FRONT_MATTER,
    }).then(({ frontMatter }) => {
        const dateMatch = file.match(DATE_FILENAME_REGEX);
        const date = dateMatch
            ? new Date(
                  Date.UTC(
                      Number(dateMatch[1]),
                      Number(dateMatch[2]) - 1,
                      Number(dateMatch[3]),
                  ),
              )
            : undefined;

        const slug = frontMatter.slug ?? file.replace(/\.mdx?$/, "");

        return {
            permalink: normalizeUrl([routeBasePath, slug]),
            title: frontMatter.title,
            description: frontMatter.description,
            tags: (frontMatter.tags ?? []).map(String),
            date,
        };
    });
}

// Independently reads blog frontmatter from disk rather than depending on
// the blog plugin's internal `allContent` shape/timing, which changes across
// Docusaurus versions and isn't part of its stable API.
module.exports = function relatedPostsPlugin(context, options) {
    const { blogPath = "blog", routeBasePath = "/blog", relatedCount = 3 } =
        options ?? {};
    const blogDir = path.join(context.siteDir, blogPath);

    return {
        name: "openpanel-related-posts-plugin",
        async loadContent() {
            const files = fs
                .readdirSync(blogDir)
                .filter((file) => /\.mdx?$/.test(file));

            return Promise.all(
                files.map((file) => loadPost(blogDir, routeBasePath, file)),
            );
        },
        async contentLoaded({ content: posts, actions }) {
            const dateFormatter = new Intl.DateTimeFormat("en", {
                year: "numeric",
                month: "long",
                day: "numeric",
                timeZone: "UTC",
            });

            const relatedByPermalink = {};

            posts.forEach((post) => {
                const related = posts.filter(
                    (candidate) =>
                        candidate.permalink !== post.permalink &&
                        candidate.tags.some((tag) =>
                            post.tags.includes(tag),
                        ),
                );

                const picked = getMultipleRandomElements(
                    related,
                    relatedCount,
                );

                relatedByPermalink[post.permalink] = picked.map((p) => ({
                    title: p.title,
                    description: p.description,
                    permalink: p.permalink,
                    date: p.date,
                    formattedDate: p.date
                        ? dateFormatter.format(p.date)
                        : undefined,
                }));
            });

            actions.setGlobalData({ relatedByPermalink });
        },
    };
};
