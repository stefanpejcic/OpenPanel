import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const License: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="LICENSE | OpenPanel">
                <html data-page="support" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>End User License Agreement (EULA)</h1>
                    <p><strong>Last updated: 05.05.2025</strong></p>

                    <p>
                        This End User License Agreement ("Agreement") is a legal agreement between you ("Licensee") 
                        and OpenPanel ("Licensor") for the use of the software applications known as OpenAdmin 
                        and OpenPanel (User Interface) (collectively, the "Software").
                    </p>

                    <p>
                        By installing, copying, or otherwise using the Software, you agree to be bound by the terms of this Agreement.
                    </p>

                    <h2>1. License Grant</h2>
                    <p>
                        Licensor grants you a non-exclusive, non-transferable, non-sublicensable license to use the Software for internal purposes only.
                        This license does not grant you the right to:
                    </p>
                    <ul>
                        <li>Distribute, resell, or share the Software.</li>
                        <li>Modify, reverse engineer, decompile, or disassemble the Software.</li>
                        <li>Use the Software for any commercial offering without Licensor's written consent.</li>
                    </ul>

                    <h2>2. Ownership</h2>
                    <p>
                        The Software is licensed, not sold. Licensor retains all rights, titles, and interests in and to the Software, 
                        including all copyrights, trademarks, and intellectual property.
                    </p>

                    <h2>3. Restrictions</h2>
                    <p>You shall not:</p>
                    <ul>
                        <li>Copy or distribute the Software or any portion thereof.</li>
                        <li>Remove or alter any proprietary notices or labels.</li>
                        <li>Use the Software in any manner that violates applicable laws or regulations.</li>
                    </ul>

                    <h2>4. Termination</h2>
                    <p>
                        This Agreement is effective until terminated. Licensor may terminate this Agreement at any time if you breach any term. 
                        Upon termination, you must destroy all copies of the Software in your possession.
                    </p>

                    <h2>5. No Warranty</h2>
                    <p>
                        The Software is provided "as is" without warranty of any kind. Licensor disclaims all warranties, express or implied, 
                        including fitness for a particular purpose and non-infringement.
                    </p>

                    <h2>6. Limitation of Liability</h2>
                    <p>
                        In no event shall Licensor be liable for any damages (including lost profits or data) arising out of the use or inability 
                        to use the Software, even if Licensor has been advised of the possibility of such damages.
                    </p>

                    <h2>7. Governing Law</h2>
                    <p>
                        This Agreement shall be governed by and construed in accordance with the laws of your jurisdiction, without regard 
                        to its conflict of law principles.
                    </p>

                    <h2>8. Entire Agreement</h2>
                    <p>
                        This Agreement constitutes the entire agreement between the parties concerning the Software and supersedes all prior agreements.
                    </p>

                    <p>
                        <strong>OpenPanel</strong><br />
                        Website: <a href="https://openpanel.com">https://openpanel.com</a><br />
                        Contact: <a href="mailto:info@openpanel.com">info@openpanel.com</a>
                    </p>
                </div>

                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default License;
