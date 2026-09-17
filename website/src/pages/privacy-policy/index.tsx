// Import React and necessary components from Docusaurus and other packages
import React from 'react';
import Head from '@docusaurus/Head';
import { BlogFooter } from '@site/src/refine-theme/blog-footer';
import { CommonHeader } from '@site/src/refine-theme/common-header';
import { CommonLayout } from '@site/src/refine-theme/common-layout';
import clsx from 'clsx';

// Import the Markdown file as a React component
import PrivacyContent from '@site/src/pages/privacy-policy/privacy.md';

const PrivacyPolicy: React.FC = () => {
    return (
        <CommonLayout
            description="Read the OpenPanel privacy policy to learn what data we collect, how it's used, and how we protect it."
            image="/img/og/privacy-policy.png"
        >
            <Head>
                <title>Privacy Policy | OpenPanel</title>
                <meta property="og:title" content="Privacy Policy | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="privacy_policy" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    {/* Render the imported Markdown content as a component */}
                    <PrivacyContent />
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default PrivacyPolicy;
