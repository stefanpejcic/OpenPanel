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
        <CommonLayout>
            <Head title="Privacy Policy | OpenPanel">
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
