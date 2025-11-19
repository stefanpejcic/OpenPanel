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
                    <p>Below is a glimpse of the features planned for OpenPanel.</p>
                    <p>We share this roadmap to give our users a clear view of our future direction and to encourage feedback on which features resonate the most (or least).</p>
                    <ul>
                        <li>Autoinstallers for popular CMS/scripts.</li>
                        <li>Additional options for WP Manager: clone, staging, scheduled backups..</li>
                         <li>ImunifyAV page for end-users.</li>
                        <li>Option to store emails on a remote server.</li>
                        <li>Ongoing improvements in container isolation.</li>
                    </ul>
                    <p>View all planned features on:<strong><a href="https://github.com/users/stefanpejcic/projects/2/views/4" target="_blank" rel="noopener noreferrer">Github</a></strong>.</p>
                    
                    <h3>Request a Feature</h3>
                    <p>Submit an idea on <strong><a href="https://github.com/stefanpejcic/OpenPanel/discussions/new?category=ideas" target="_blank" rel="noopener noreferrer">Github Discussions</a></strong>.</p>

                    <h3>Vote for a Feature</h3>
                    <p>Vote for existing feature requests on <strong><a href="https://github.com/stefanpejcic/OpenPanel/discussions/categories/ideas" target="_blank" rel="noopener noreferrer">Github Discussions</a></strong>.</p>
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
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Roadmap;
