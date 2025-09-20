import React from "react";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import Head from "@docusaurus/Head";
import clsx from "clsx";

const NetworkTablePage: React.FC = () => {
  return (
    <>
      <Head title="Network Pricing | OpenPanel Project">
        <html data-page="network-table" />
      </Head>

      <div className="refine-prose">
        <CommonHeader hasSticky={true} />

        <div
          className={clsx(
            "max-w-[912px] w-full mx-auto py-16",
            "text-center text-gray-800 dark:text-gray-200"
          )}
        >
          <p>
            If you have your own IP ranges like /24 or any other then you can get OpenPanel Enterprise for all IP’s in the range at the best price.
          </p>
          <p>            
            <a href="https://my.openpanel.com/index.php?rp=/store/partners/noc" target="_blank">OpenPanel NOC Partner package</a> prices based on the Network Prefix Length (see the table bellow).
          </p>

          <div className="overflow-x-auto mt-8">
            <table className="border border-gray-300 w-full text-left">
              <thead>
                <tr className="bg-gray-100 dark:bg-gray-800">
                  <th className="border px-4 py-2">Network Bits</th>
                  <th className="border px-4 py-2">Subnet Mask</th>
                  <th className="border px-4 py-2">Number of Subnets</th>
                  <th className="border px-4 py-2">Number of Hosts</th>
                  <th className="border px-4 py-2">EUR/Monthly</th>
                </tr>
              </thead>
              <tbody>
                <tr><td>/8</td><td>255.0.0.0</td><td>0</td><td>16777214</td><td></td></tr>
                <tr><td>/9</td><td>255.128.0.0</td><td>2 (0)</td><td>8388606</td><td></td></tr>
                <tr><td>/10</td><td>255.192.0.0</td><td>4 (2)</td><td>4194302</td><td></td></tr>
                <tr><td>/11</td><td>255.224.0.0</td><td>8 (6)</td><td>2097150</td><td></td></tr>
                <tr><td>/12</td><td>255.240.0.0</td><td>16 (14)</td><td>1048574</td><td></td></tr>
                <tr><td>/13</td><td>255.248.0.0</td><td>32 (30)</td><td>524286</td><td></td></tr>
                <tr><td>/14</td><td>255.252.0.0</td><td>64 (62)</td><td>262142</td><td>€262142.00</td></tr>
                <tr><td>/15</td><td>255.254.0.0</td><td>128 (126)</td><td>131070</td><td>€131070.00</td></tr>
                <tr><td>/16</td><td>255.255.0.0</td><td>256 (254)</td><td>65534</td><td>€65534.00</td></tr>
                <tr><td>/17</td><td>255.255.128.0</td><td>512 (510)</td><td>32766</td><td>€32766.00</td></tr>
                <tr><td>/18</td><td>255.255.192.0</td><td>1024 (1022)</td><td>16382</td><td>€16382.00</td></tr>
                <tr><td>/19</td><td>255.255.224.0</td><td>2048 (2046)</td><td>8190</td><td>€8190.00</td></tr>
                <tr><td>/20</td><td>255.255.240.0</td><td>4096 (4094)</td><td>4094</td><td>€4094.00</td></tr>
                <tr><td>/21</td><td>255.255.248.0</td><td>8192 (8190)</td><td>2046</td><td>€2046.00</td></tr>
                <tr><td>/22</td><td>255.255.252.0</td><td>16384 (16382)</td><td>1022</td><td>€1022.00</td></tr>
                <tr><td>/23</td><td>255.255.254.0</td><td>32768 (32766)</td><td>510</td><td>€510.00</td></tr>
                <tr><td>/24</td><td>255.255.255.0</td><td>65536 (65534)</td><td>254</td><td>€254.00</td></tr>
                <tr><td>/25</td><td>255.255.255.128</td><td>131072 (131070)</td><td>126</td><td>€126.00</td></tr>
                <tr><td>/26</td><td>255.255.255.192</td><td>262144 (262142)</td><td>62</td><td>€62.00</td></tr>
                <tr><td>/27</td><td>255.255.255.224</td><td>524288 (524286)</td><td>30</td><td>€30.00</td></tr>
                <tr><td>/28</td><td>255.255.255.240</td><td>1048576 (1048574)</td><td>14</td><td>€14.00</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </>
  );
};

export default function NetworkTable() {
  return (
    <CommonLayout>
      <NetworkTablePage />
    </CommonLayout>
  );
}
