import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const PrivacyPolicy: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="Privacy Policy | OpenPanel">
                <html data-page="privacy_policy" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Privacy Policy</h1>
                    <p>Last updated: April 04, 2023</p>
                    <p>
                        This Privacy Policy describes Our policies and
                        procedures on the collection, use and disclosure of Your
                        information when You use the Service and tells You about
                        Your privacy rights and how the law protects You.
                    </p>
                    <p>
                        We use Your Personal data to provide and improve the
                        Service. By using the Service, You agree to the
                        collection and use of information in accordance with
                        this Privacy Policy.
                    </p>
                    <h1>Interpretation and Definitions</h1>
                    <h2 style={{ marginTop: 0 }}>Interpretation</h2>
                    <p>
                        The words of which the initial letter is capitalized
                        have meanings defined under the following conditions.
                        The following definitions shall have the same meaning
                        regardless of whether they appear in singular or in
                        plural.
                    </p>
                    <p>
                        We will let You know via email and/or a prominent notice
                        on Our Service, prior to the change becoming effective
                        and update the &quot;Last updated&quot; date at the top
                        of this Privacy Policy.
                    </p>
                    <p>
                        You are advised to review this Privacy Policy
                        periodically for any changes. Changes to this Privacy
                        Policy are effective when they are posted on this page.
                    </p>
                    <h1>Contact Us</h1>
                    <p>
                        If you have any questions about this Privacy Policy, You
                        can contact us:
                    </p>
                    <ul>
                        <li>By email: info@refine.dev</li>
                    </ul>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};
export default PrivacyPolicy;
