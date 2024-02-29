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
        <p>OpenPanel is committed to protecting your privacy and any personal information you share with us. Accordingly, we have developed this privacy policy in order for you to understand how we collect, use, communicate, disclose and otherwise make use of personal information.</p>
        <p>We have outlined our privacy policy below. It is recommended that you read this privacy policy carefully. If you have additional questions or would like further information on this topic, please feel free to contact us.</p>

        <h2>ABOUT OPENPANEL</h2>
        <p>OpenPanel is the controller for the processing of your personal data, as described in this privacy policy. Our company particulars are:</p>
        <address>
            OpenPanel<br>
            IJsbaanpad 2<br>
            1076 CV Amsterdam<br>
            The Netherlands<br>
            <br>
            Privacy Protection Team: <a href="mailto:privacy@openpanel.co">privacy@openpanel.co</a>
        </address>

        <h2>TO WHOM DOES THIS PRIVACY POLICY APPLY?</h2>
        <p>This privacy policy applies when you visit and use the OpenPanel website and OpenPanel products and get in touch with us through the website or otherwise. This privacy policy does not apply to the content and data processed, stored, or hosted by our customers using OpenPanel services.</p>

        <h2>HOW DO WE USE YOUR PERSONAL DATA?</h2>
        <p>Below is an overview of the purposes for which OpenPanel processes your personal data. In this overview, OpenPanel indicates the personal data OpenPanel uses for a specific purpose, what the legal basis is for processing these data, and how long OpenPanel stores the data. To keep things organised, we have grouped everything by type of data stream.</p>
        <h3>Legal bases: explanatory note</h3>
        <ul>
            <li>Article 6.1(a) GDPR: This means that we process your personal data based on your consent.</li>
            <li>Article 6.1(b) GDPR: This means that we need your personal data to enter into a contract with you or to take steps at your request prior to entering into a contract.</li>
            <li>Article 6.1(f) GDPR: This means that we need your personal data for the purposes of a legitimate interest pursued by us or by a third party.</li>
        </ul>
        <h3>Customers</h3>
        <p>When you enter into an agreement with OpenPanel, we have to collect and process certain personal data from you. We need your name, address information, telephone number, payment information and your email address to process your order. If you have registered for our proactive monitoring tool, we will use your email address to inform you about disruption notifications from our data centers.</p>
        <p>When paying by credit card we do not collect and store any payment information such as credit card numbers or verification codes. You disclose these only directly to the respective payment service provider. Additionally, If you pay via credit card we may ask you for:</p>
        <ul>
            <li>Copy front credit card (last 4 digits, exp. date, and visible, all other data should be hidden).</li>
        </ul>
        <p>We delete these data as soon as you delete your user account with us or once the statutory retention period of 7 years for tax purposes expires. You can delete your user account with us by sending an e-mail to <a href="mailto:privacy@openpanel.com">privacy@openpanel.com</a> with your deletion request. The legal basis for this processing is Art. 6.1(b) GDPR.</p>
        <h3>Contact</h3>
        <p>To provide customer service, OpenPanel may collect your personal data to respond to your questions you submit through email or the contact form on the website, including helping you with any issues which may arise regarding our website and/or services. If you use the online chat function, the data processed by OpenPanel for this purpose are IP address, your name, email address and data that you enter in the contact form to or a chat conversation with OpenPanel The processing of these personal data is necessary for the purposes of the legitimate interests pursued by OpenPanel (Article 6.1(f) GDPR), namely being able to serve you efficiently and to optimize OpenPanel´s customer service. We may also contact you at a later moment about our services. Your personal data is stored for a period of 6 months after responding to your message or resolving your complaint.</p>
        <h3>Newsletters</h3>
        <p>OpenPanel wishes to inform you on the developments of its products, if you have signed up for OpenPanel’s newsletters. If, at any moment, you do not wish to receive emails of OpenPanel anymore, you can unsubscribe by using the opt-out option that is provided in every email that OpenPanel sends you. You may also unsubscribe by sending an email to <a href="mailto:privacy@openpanel.co">privacy@openpanel.co</a>.</p>
        <p>To subscribe you have to provide us with your email address. The processing thereof is necessary for the purposes of the legitimate interests pursued by OpenPanel (Article 6.1(f) GDPR), namely direct marketing. OpenPanel retains your e-mail address for as long as you are subscribed to the newsletter and not longer than 1 month after you unsubscribe.</p>
        <h3>Careers</h3>
        <p>If you apply for a position at OpenPanel, you will provide your personal data such as your first and last name, full address, email address, phone number and your CV and other attachments to your application. OpenPanel uses this information to review your application and respond to your application. The processing of these personal data is necessary for the purposes of the legitimate interests pursued by OpenPanel (Article 6.1(f) GDPR), namely recruitment. OpenPanel retains your personal data for 4 weeks after rejection or, with your consent (Article 6(a) GDPR), for a period of 1 year.</p>
        <h3>Server installations log files</h3>
        <p>OpenPanel automatically collects generated data about your installation. This information consists of your server IP address (an unique number, which makes it possible to connect your licence with server); licence activity and updates logs.</p>
        <p>OpenPanel requires this information in order for the serverice to work as optimally as possible, the processing of these personal data is therefore necessary for the purposes of the legitimate interests pursued by OpenPanel (Article 6.1(f) GDPR).</p>

        <h2>DOES OPENPANEL SHARE YOUR PERSONAL DATA?</h2>
        <p>When you make a transaction, your transaction data is sent to the relevant payment service provider in order to process your payment. For example:</p>
        <ul>
            <li>When paying by credit card, your payment is processed by Stripe Inc. (United States).</li>
        </ul>
        <p>If you request OpenPanel to register a domain on your behalf, we need to provide the required data to the corresponding registration service provider.</p>
        <p>In addition, OpenPanel will provide your personal data to external parties, such as police and government institutions if OpenPanel is obliged to do so on the basis applicable laws and / or regulations or by means of a court order or a legal verdict, or has obtained consent from you for the provision to third parties.</p>

        <h2>DOES OPENPANEL TRANSFER YOUR PERSONAL DATA OUTSIDE THE EUROPEAN ECONOMIC AREA?</h2>
        <p>To comply with EU data protection laws around international data transfer, OpenPanel only allows its service providers outside the EEA to process your personal data in accordance with a contract entered into between the OpenPanel and the service provider, incorporating the European Commission’s Standard Contractual Clauses, which ensure that adequate data protection arrangements are in place (Article 46.1(c) GDPR). For more information on where and how the relevant document may be accessed or obtained, please contact us.</p>

        <h2>WHAT ARE YOUR RIGHTS?</h2>
        <p>Under the GDPR, you have a number of rights with regard to your personal data and the processing thereof:</p>
        <ul>
            <li><strong>Right to access:</strong> You have the right to obtain confirmation as to whether or not your personal data are being processed, and, if so, access to your personal data and additional information about the processing of your personal data.</li>
            <li><strong>Right to rectification:</strong> you have the right to request the rectification of inaccurate personal data.</li>
            <li><strong>Right to be forgotten:</strong> You have the right to ask us to erase your personal data (right to be forgotten) for example if the personal data are no longer necessary in relation to the purposes for which they were collected; or the personal data have been unlawfully processed.</li>
            <li><strong>Right to restriction:</strong> You have the right to obtain restriction of processing of your personal data, for example when you have contested the accuracy of your personal data.</li>
            <li><strong>Data portability:</strong> You have the right to receive your personal data which you have provided to us in a structured, commonly used and machine-readable format and have the right to transmit those data to another controller, where the processing is based on your consent or on a contract.</li>
            <li><strong>Right to object:</strong> You have the right to object to processing of personal data which is based on our legitimate interests. We shall no longer process the personal data unless we demonstrate compelling legitimate grounds for the processing which override your interests, rights and freedoms or for the establishment, exercise or defence of legal claims; where personal data are processed for direct marketing purposes, you always have the right to object to processing of personal data for such marketing. In that case, we shall no longer process your personal data for such purposes.</li>
            <li><strong>Withdraw consent:</strong> Where the processing of your personal data is based on your consent, you have the right to withdraw your consent at any time. The withdrawal of consent does not affect the lawfulness of processing based on consent before its withdrawal.</li>
        </ul>
        <p>OpenPanel shall respond to your request without undue delay and in any event within one month of receipt of the request. That period may be extended by two further months where necessary, taking into account the complexity and number of the requests. OpenPanel shall inform you of any such extension within one month of receipt of the request, together with the reasons for the delay.</p>
        <p>If OpenPanel does not take action on your request, OpenPanel shall inform you without delay and at the latest within one month of receipt of the request of the reasons for not taking action. In that case, you also have the right to lodge a complaint with the supervisory authority and to seek a judicial remedy.</p>

        <h2>WHAT ARE COOKIES AND HOW DOES OPENPANEL USE THEM?</h2>
        <p>When using the website or app, various cookies are saved on your computer, mobile device and/or tablet, which can then be accessed. Please consult the OpenPanel cookie policy for more information about how OpenPanel uses cookies.</p>

        <h2>COMPLAINTS?</h2>
        <p>In case you suspect a breach of data protection legislation and the matter is not solved amicably between us in negotiations, you also have the right to lodge a complaint with a supervisory authority.</p>

                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default PrivacyPolicy;
