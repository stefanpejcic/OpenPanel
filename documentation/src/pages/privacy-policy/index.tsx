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
                <html data-page="support" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Privacy Policy</h1>
                       <p>
                           OpenPanel is committed to protecting your privacy and any personal information you share with us. Accordingly, we have developed this privacy policy in order for you to understand how we collect, use, communicate, disclose and otherwise make use of personal information.
                       </p>
                        <p>
                           We have outlined our privacy policy below. It is recommended that you read this privacy policy carefully. If you have additional questions or would like further information on this topic, please feel free to contact us. 
                        </p>
                     <h2>ABOUT OPENPANEL</h2>
                        <p>
                            OpenPanel is the controller for the processing of your personal data, as described in this privacy policy. Our company particulars are:

                            OpenPanel<br>
                            IJsbaanpad 2<br>
                            1076 CV Amsterdam<br>
                            The Netherlands<br>
                            <br>
                            Privacy Protection Team:
                                <a href="mailto:privacy@openpanel.co">privacy@openpanel.co</a>
                        </p>
                        <h2>TO WHOM DOES THIS PRIVACY POLICY APPLY?</h2>
                        <p>
                            This privacy policy applies when you visit and use the OpenPanel website and OpenPanel products and get in touch with us through the website or otherwise. This privacy policy does not apply to the content and data processed, stored, or hosted by our customers using OpenPanel services.
                        </p>                    
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default PrivacyPolicy;
