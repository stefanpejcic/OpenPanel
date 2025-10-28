import React from 'react';
import OriginalRoot from '@theme-original/Root';
import { useEffect } from 'react';

export default function Root(props) {
  useEffect(() => {
    // LinkedIn Partner ID
    const partnerId = '8893345';
    window._linkedin_data_partner_ids = window._linkedin_data_partner_ids || [];
    window._linkedin_data_partner_ids.push(partnerId);

    // Create LinkedIn tracking script
    const script = document.createElement('script');
    script.type = 'text/javascript';
    script.async = true;
    script.src = 'https://snap.licdn.com/li.lms-analytics/insight.min.js';
    document.body.appendChild(script);

    // Add noscript fallback
    const noscript = document.createElement('noscript');
    noscript.innerHTML = `<img height="1" width="1" style="display:none;" alt="" src="https://px.ads.linkedin.com/collect/?pid=${partnerId}&fmt=gif" />`;
    document.body.appendChild(noscript);
  }, []);

  return <OriginalRoot {...props} />;
}
