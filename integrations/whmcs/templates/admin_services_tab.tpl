<div class="">
    <!-- Disk Usage Panel -->
    {if $disk}
    <div class="panel panel-default">
        <div class="panel-heading"><strong>Storage</strong></div>
        <div class="panel-body">
            {foreach $diskItems as $item}
            <div class="mb-2">
                <div class="text-muted">{$item.label} <span class="pull-right">{$item.value} / {$item.max} ({$item.percent}%)</span></div>
                <div class="progress">
                    <div class="progress-bar" role="progressbar" style="width: {$item.percent}%;">{$item.percent}%</div>
                </div>
            </div>
            {/foreach}
        </div>
    </div>
    {/if}

    <!-- Domains & Sites Panel -->
    <div class="panel panel-default">
        <div class="panel-heading"><strong>Domains</strong> 
            <span class="text-muted">(Total Domains: {$totalDomains}, Total Sites: {$totalSites})</span>
        </div>
        <div class="panel-body">
            {if $domains}
                {foreach $domains as $domain}
                    <div class="border rounded p-2 mb-2">
                        <div class="fw-bold">{$domain.domain_url}</div>
                        <div class="text-muted small">PHP: {$domain.php_version|default:'—'}, Docroot: {$domain.docroot|default:'—'}</div>
                        {if $domainSites[$domain.domain_id]}
                        <ul class="small mb-0">
                            {foreach $domainSites[$domain.domain_id] as $site}
                                <li>{$site.site_name|default:'—'} ({$site.type|default:'—'}, Admin: {$site.admin_email|default:'—'}, Created: {$site.created_date|date_format:'%Y-%m-%d'|default:'—'})</li>
                            {/foreach}
                        </ul>
                        {/if}
                    </div>
                {/foreach}
            {else}
                <span class="text-muted">None</span>
            {/if}
        </div>
    </div>

    <!-- Plan Info Panel -->
    <div class="panel panel-default">
        <div class="panel-heading"><strong>Plan Limits</strong></div>
        <div class="panel-body">
            {if $plan}
            <table class="table table-sm table-bordered mb-0">
                <tbody>
                {foreach $planFields as $key => $label}
                    <tr><th>{$label}</th><td>{$plan[$key]|default:'—'}</td></tr>
                {/foreach}
                </tbody>
            </table>
            {else}
                <span class="text-muted">—</span>
            {/if}
        </div>
    </div>

    <!-- User Info Panel -->
    <div class="panel panel-default">
        <div class="panel-heading"><strong>Account Information</strong></div>
        <div class="panel-body">
            <table class="table table-sm table-bordered mb-0">
                <tbody>
                    <tr><th>Email</th><td>{$user.email|default:'—'}</td></tr>
                    <tr><th>Docker context</th><td>{$user.server|default:'—'}</td></tr>
                    <tr><th>Owned by Reseller</th><td>{if $user.owner}{$user.owner}{else}<span class="text-muted">No</span>{/if}</td></tr>
                    <tr><th>2FA Enabled</th><td>{if $user.twofa_enabled}<span class="badge bg-success">Yes</span>{else}<span class="badge bg-secondary">No</span>{/if}</td></tr>
                    <tr><th>OpenPanel account created</th><td>{$user.registered_date|date_format:'%Y-%m-%d %H:%M'|default:'—'}</td></tr>
                </tbody>
            </table>
        </div>
    </div>
</div>
