import React from "react";
import Head from "@docusaurus/Head";
import { PageMetadata } from "@docusaurus/theme-common";
import { useDoc } from "@docusaurus/theme-common/internal";

// Keep in sync with permalinkToOgImagePath() and OG_WIDTH/OG_HEIGHT in plugins/og-images.ts
function permalinkToOgImagePath(permalink) {
    const clean = permalink.replace(/\/+$/, "") || "/index";
    return `/img/og${clean}.png`;
}

const OG_WIDTH = 1200;
const OG_HEIGHT = 630;

export default function DocItemMetadata() {
    const { metadata, frontMatter, assets } = useDoc();

    const hasCustomImage = Boolean(assets.image ?? frontMatter.image);
    const image =
        assets.image ?? frontMatter.image ?? permalinkToOgImagePath(metadata.permalink);

    return (
        <>
            <PageMetadata
                title={metadata.title}
                description={metadata.description}
                keywords={frontMatter.keywords}
                image={image}
            />
            {!hasCustomImage && (
                <Head>
                    <meta property="og:image:width" content={String(OG_WIDTH)} />
                    <meta property="og:image:height" content={String(OG_HEIGHT)} />
                    <meta property="og:image:type" content="image/png" />
                </Head>
            )}
        </>
    );
}
