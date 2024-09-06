import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const PrivacyPolicy: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="Support | OpenPanel">
                <html data-page="support" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Support</h1>
                    <p><strong>Don't worry - we're here to help</strong></p>
                    <p>Make the most of your OpenPanel deployments, find quick solutions to common problems, and get direct assistance from the developers of OpenPanel and OpenAdmin. Usually, the quickest way to get answers is as follows:</p>
                
                    <h3>Step 1: Search for similar questions</h3>
                    <p>If your question has been asked before, you might be able to find an answer in minutes and get your problem knocked out before we even know it exists.</p>
                
                    <ul>
                        <li><a href="/search">Search Documentation</a></li>
                        <li><a href="https://community.openpanel.com">Search Forums</a></li>
                    </ul>
                
                    <p>If a search does not provide good answers, proceed to step 2.</p>
                                
                    <h3>Step 2a: Open a support ticket (Enterprise Edition)</h3>
                    <p>If you need fast and reliable support, open a premium ticket by sending us a new ticket.</p>
                    <ul>
                        <li><a href="https://my.openpanel.com/submitticket.php?step=2&deptid=1" target="_blank">Submit a ticket</a></li>
                    </ul>
                
                    <h3>Step 2b: Post on forums (Community Edition)</h3>
                    <p>If you can wait for answers from our community and don't want to spend any money, ask your question in the forums. Every member of the OpenPanel staff also monitors the forums and tries to help out whenever we can. We generally prioritize paid support queries (we've got bills to pay, after all), but our forums are very active, and our community is knowledgeable.</p>
                    <ul>
                        <li><a href="https://community.openpanel.com">Ask the Community</a></li>
                    </ul>
                                
                    <h3>Step 3: Found a bug?</h3>
                    <p>If you've found a bug in OpenPanel or OpenAdmin, Community or Enterprise edition, <a
                            href="https://github.com/stefanpejcic/openpanel/issues/new?assignees=&labels=Bug&projects=&template=1_Bug_report.yaml">please
                            file a bug</a> in the relevant GitHub issue tracker for the project you're reporting on or <a
                            href="https://community.openpanel.com">post to the forum</a> or report to the appropriate <code>security@</code>
                        email address if reporting security issues. It benefits everyone when you let us know about bugs you find in our
                        software.</p>

                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default PrivacyPolicy;
