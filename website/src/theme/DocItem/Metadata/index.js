import React from "react";
import { PageMetadata } from "@docusaurus/theme-common";
import { useDoc } from "@docusaurus/theme-common/internal";

// Keep in sync with permalinkToOgImagePath() in plugins/og-images.ts
function permalinkToOgImagePath(permalink) {
    const clean = permalink.replace(/\/+$/, "") || "/index";
    return `/img/og${clean}.png`;
}

export default function DocItemMetadata() {
    const { metadata, frontMatter, assets } = useDoc();

    const image =
        assets.image ?? frontMatter.image ?? permalinkToOgImagePath(metadata.permalink);

    return (
        <PageMetadata
            title={metadata.title}
            description={metadata.description}
            keywords={frontMatter.keywords}
            image={image}
        />
    );
}
