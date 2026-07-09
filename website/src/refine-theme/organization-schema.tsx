import React from "react";

// Sitewide Organization + WebSite structured data (https://schema.org/Organization),
// rendered once via src/theme/Root so it applies to every page, including docs/blog.
export const OrganizationSchema: React.FC = () => {
    const jsonLd = {
        "@context": "https://schema.org",
        "@graph": [
            {
                "@type": "Organization",
                name: "OpenPanel",
                url: "https://openpanel.com",
                logo: "https://openpanel.com/img/svg/openpanel_logo.svg",
                sameAs: [
                    "https://github.com/stefanpejcic/openpanel",
                    "https://discord.com/invite/7bNY8fANqF",
                    "https://www.reddit.com/r/openpanelco/",
                    "https://x.com/openpanel",
                    "https://www.linkedin.com/company/openpanel/",
                ],
            },
            {
                "@type": "WebSite",
                name: "OpenPanel",
                url: "https://openpanel.com",
            },
        ],
    };

    return (
        <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{
                __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c"),
            }}
        />
    );
};
