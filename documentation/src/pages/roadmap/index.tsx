import React, { useEffect, useState } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Roadmap: React.FC = () => {
    const [milestones, setMilestones] = useState<any[]>([]);

    useEffect(() => {
        const fetchMilestones = async () => {
            try {
                const response = await fetch("https://api.github.com/repos/stefanpejcic/openpanel/milestones");
                const data = await response.json();
                setMilestones(data);
            } catch (error) {
                console.error("Error fetching milestones:", error);
            }
        };

        fetchMilestones();
    }, []);

    const calculateProgress = (milestone) => {
        const totalIssues = milestone.open_issues + milestone.closed_issues;
        const progress = (milestone.closed_issues / totalIssues) * 100;
        return progress.toFixed(2);
    };

    const formatDueDate = (dueDate) => {
        const options = { year: 'numeric', month: 'long', day: 'numeric' };
        return new Date(dueDate).toLocaleDateString(undefined, options);
    };

    return (
        <CommonLayout>
            <Head title="ROADMAP | OpenPanel">
                <html data-page="roadmap" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Roadmap</h1>
                    <p>The following is a list of features in the OpenPanel roadmap.</p>
                    <p>We list these here so that our users can get a better idea of where we are going and can give us feedback on which of these they like best (or least).</p>
                    <p>Here is a short list; the Github repository has many more.</p>
                    <ul>
                        <li>Support for more Linux distros.</li>
                        <li>More improvements to cluster mode and OpenAdmin.</li>
                        <li>E-Mail support.</li>
                        <li>Continued improvements to container isolation.</li>
                    </ul>
                    <h3>Planned releases</h3>
                    <ul>
                        {milestones.map((milestone) => (
                            <li key={milestone.id}>
                                <strong><a href={milestone.html_url} target="_blank" rel="noopener noreferrer">{milestone.title}</a></strong>
                                <p>{milestone.description}</p>
                                <p>Scheduled release: <strong>{formatDueDate(milestone.due_on)}</strong></p>
                                <p>Progress: {calculateProgress(milestone)}%</p>
                                
                                {/* Progress bar */}
                                <div style={{ width: '100%', height: '8px', backgroundColor: '#ddd', borderRadius: '4px', marginTop: '8px', overflow: 'hidden' }}>
                                    <div style={{ height: '100%', backgroundColor: '#4caf50', width: `${calculateProgress(milestone)}%`, transition: 'width 0.3s ease-in-out' }}></div>
                                </div>
                            </li>
                        ))}
                    </ul>

                    <h3>About</h3>
                    <p>At OPENPANEL we believe in the power of collaboration and value the feedback of our community. This roadmap is a testament to our dedication to open communication and customer-centric development. As we embark on this journey together, we invite you to explore the exciting initiatives that lie ahead and join us in shaping the future of hosting.</p>
                    <p>Discover the roadmap that outlines our commitment to providing a robust, user-friendly hosting panel that not only meets but exceeds your expectations. Your insights and suggestions are invaluable to us, and we look forward to the shared successes that await on this path of innovation. Thank you for being an integral part of our growing community!</p>
                    
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Roadmap;
