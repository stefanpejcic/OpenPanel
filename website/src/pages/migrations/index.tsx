// Import React and necessary components from Docusaurus and other packages
import React from 'react';
import Head from '@docusaurus/Head';
import { BlogFooter } from '@site/src/refine-theme/blog-footer';
import { CommonHeader } from '@site/src/refine-theme/common-header';
import { CommonLayout } from '@site/src/refine-theme/common-layout';
import clsx from 'clsx';

// Import the Markdown file as a React component
import MigrationsContent from '@site/src/pages/migrations/migrations.md';

const Migrations: React.FC = () => {
    return (
        <CommonLayout
            description="Free migration from cPanel, Plesk, DirectAdmin and CyberPanel to OpenPanel for annual Enterprise license holders - up to 20 accounts and 500 GB included."
            image="/img/og/migrations.png"
        >
            <Head>
                <title>Free Migration to OpenPanel | OpenPanel</title>
                <meta property="og:title" content="Free Migration to OpenPanel | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="migrations" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    {/* Render the imported Markdown content as a component */}
                    <MigrationsContent />
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Migrations;
