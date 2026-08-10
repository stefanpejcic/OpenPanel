import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const License: React.FC = () => {
    return (
        <CommonLayout description="The OpenPanel End User License Agreement (EULA) covering license grant, restrictions, ownership, and purchase terms.">
            <Head title="LICENSE | OpenPanel">
                <html data-page="license" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>End User License Agreement (EULA)</h1>
                    <p><strong>Last updated: 10.08.2026</strong></p>

                    <p>
                        This End User License Agreement ("Agreement") is a legal agreement between you ("Licensee")
                        and OpenPanel, LLC. ("Licensor") for the use of the software applications known as OpenAdmin
                        and OpenPanel (User Interface) (collectively, the "Software").
                    </p>

                    <p>
                        By installing, copying, forking, modifying, or otherwise using the Software, you agree to be bound by the terms of this Agreement.
                    </p>

                    <h2>1. License Grant</h2>
                    <p>
                        Licensor grants you a worldwide, royalty-free, non-exclusive license to install, run, copy, modify, and fork the
                        Software for personal, educational, evaluation, non-commercial, and internal business purposes, including
                        self-hosting it in production for your own organization's use (subject to the Community Edition's published
                        limits — see <a href="/enterprise">openpanel.com/enterprise</a>).
                    </p>
                    <p>
                        This license does not grant you the right to:
                    </p>
                    <ul>
                        <li>Use the Software, or any fork or modified version of it, to provide a commercial offering to third parties — including hosting it as a paid or ad-supported service, reselling it, or operating a hosting or reseller business built on it — without a valid OpenPanel Enterprise license or Licensor's separate written consent.</li>
                        <li>Use OpenPanel's name, logo, or other trademarks to imply endorsement of, or affiliation with, any fork or derivative work.</li>
                    </ul>

                    <h2>2. Ownership</h2>
                    <p>
                        Licensor retains all rights, title, and interest in and to the Software, including all copyrights, trademarks,
                        and intellectual property. This Agreement does not transfer ownership of the Software or any component of it.
                    </p>

                    <h2>3. Forks and Redistribution</h2>
                    <p>If you publish a fork, mirror, or modified version of the Software, you must:</p>
                    <ul>
                        <li>Clearly and prominently identify it as a fork or modified version of OpenPanel, with a visible link back to the original project (<a href="https://github.com/stefanpejcic/OpenPanel">github.com/stefanpejcic/OpenPanel</a>).</li>
                        <li>Retain all copyright, trademark, and proprietary notices included in the Software.</li>
                    </ul>
                    <p>
                        Commercial use of a fork or modified version is subject to the same licensing requirements as the unmodified Software (see Section 1).
                    </p>

                    <h2>4. Restrictions</h2>
                    <p>You shall not:</p>
                    <ul>
                        <li>Circumvent, disable, or tamper with the Software's license-key verification or feature-gating mechanism in order to access Enterprise features without a valid license.</li>
                        <li>Remove or alter any proprietary or trademark notices except as permitted under Section 3.</li>
                        <li>Use the Software in any manner that violates applicable laws or regulations.</li>
                    </ul>

                    <h2>5. Termination</h2>
                    <p>
                        This Agreement is effective until terminated. Licensor may terminate this Agreement, and revoke any active
                        Enterprise license key, at any time if you breach any term. Upon termination, you must stop any commercial
                        use of the Software until you obtain a valid license.
                    </p>

                    <h2>6. No Warranty</h2>
                    <p>
                        The Software is provided "as is" without warranty of any kind. Licensor disclaims all warranties, express or implied,
                        including fitness for a particular purpose and non-infringement.
                    </p>

                    <h2>7. Limitation of Liability</h2>
                    <p>
                        In no event shall Licensor be liable for any damages (including lost profits or data) arising out of the use or inability
                        to use the Software, even if Licensor has been advised of the possibility of such damages.
                    </p>

                    <h2>8. Governing Law</h2>
                    <p>
                        This Agreement shall be governed by the laws of the State of Delaware and the United States of America.
                    </p>

                    <h2>9. Purchases & Refunds</h2>
                    <p>
                        Purchases of OpenPanel Enterprise licenses are subject to our Refund Policy available at https://openpanel.com/refund-policy
                    </p>

                    <p>
                        Refunds are only available within <strong>7 days</strong> of purchase and solely <strong>for licenses that have not been activated</strong>.
                        Once a license key is activated on any server or environment, all sales are final.
                    </p>
                    <p>
                        We offer a free Community Edition and a 30-day Enterprise trial to evaluate the software before purchase.
                    </p>

                    <h2>10. Entire Agreement</h2>
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
