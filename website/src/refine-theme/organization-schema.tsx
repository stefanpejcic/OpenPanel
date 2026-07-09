import React from "react";

// Renders Organization structured data (https://schema.org/Organization) so
// search engines can associate the site with the OpenPanel brand/logo.
export const OrganizationSchema: React.FC = () => {
    const jsonLd = {
        "@context": "https://schema.org",
        "@type": "Organization",
        name: "OpenPanel",
        url: "https://openpanel.com",
        logo: "https://openpanel.com/img/openpanel_social.png",
        sameAs: [
            "https://github.com/stefanpejcic/OpenPanel",
            "https://www.linkedin.com/company/openpanel/",
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
