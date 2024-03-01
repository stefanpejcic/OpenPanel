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
                    <p>Last updated: March 01, 2024</p>
                    <p>
                        OpenPanel is committed to protecting your privacy and any personal information you share with us. Accordingly, we have developed this privacy policy in order for you to understand how we collect, use, communicate, disclose and otherwise make use of personal information.
                    </p>
                    <p>
                        We have outlined our privacy policy below. It is recommended that you read this privacy policy carefully. If you have additional questions or would like further information on this topic, please feel free to contact us.
                    </p>
                    </p>
                    <h2 style={{ marginTop: 0 }}>ABOUT OPENPANEL</h2>
                    <p>
                        OpenPanel is the controller for the processing of your personal data, as described in this privacy policy. Our company particulars are:
                    </p>
                    <address>
                        OpenPanel<br>
                        IJsbaanpad 2<br>
                        1076 CV Amsterdam<br>
                        The Netherlands<br>
                        <br>
                        Privacy Protection Team: <a href="mailto:privacy@openpanel.co">privacy@openpanel.co</a>
                    </address>
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
