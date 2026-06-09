import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description: "Privacy Policy for Vats Apartment - how we collect, use, and protect your personal information.",
  alternates: { canonical: 'https://vatsapartment.com/privacy' },
  robots: { index: true, follow: true },
};

export default function PrivacyPage() {
  return (
    <main className="container-max gutter py-12 md:py-16 font-sans max-w-4xl mx-auto">
      <h1 className="text-3xl md:text-4xl font-bold text-on-surface mb-2">
        Privacy Policy
      </h1>
      <p className="text-on-surface-variant text-sm mb-10">
        Last updated: June 2026
      </p>

      <div className="prose prose-neutral max-w-none space-y-8 text-on-surface text-sm md:text-base leading-relaxed">

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">1. Introduction</h2>
          <p className="text-on-surface-variant">
            Vats Apartment (&quot;we,&quot; &quot;our,&quot; or &quot;us&quot;) is committed to protecting your privacy. This
            Privacy Policy explains how we collect, use, disclose, and safeguard your information when you
            visit our website <strong>vatsapartment.com</strong>, use our services, or interact with us.
            By using our website or services, you consent to the practices described in this policy.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">2. Information We Collect</h2>
          <p className="text-on-surface-variant mb-3">
            We collect the following types of information to provide and improve our accommodation services:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li><strong>Personal Identification Information:</strong> Full name, phone number, email address, date of birth, government-issued ID proof, college ID, and photograph &mdash; collected when you book a room, create a tenant account, or submit an inquiry.</li>
            <li><strong>Contact Information:</strong> Phone number, email address, and WhatsApp contact details for communication regarding bookings and services.</li>
            <li><strong>Payment Information:</strong> Rent payment history, security deposit records, electricity bill payments, and transaction details. We do <em>not</em> store full credit card or bank account numbers on our servers.</li>
            <li><strong>Usage Data:</strong> Browser type, device information, IP address, pages visited, time spent on pages, and other diagnostic data to improve our website experience.</li>
            <li><strong>Communication Data:</strong> Records of your correspondence with us via email, phone, WhatsApp, or contact forms.</li>
            <li><strong>Verification Documents:</strong> Government ID proof, college/university ID, and photographs required for tenant verification as per local regulations.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">3. How We Use Your Information</h2>
          <p className="text-on-surface-variant mb-3">
            We use the collected information for the following purposes:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>To process and manage room bookings, reservations, and tenant registrations.</li>
            <li>To verify tenant identity and comply with local accommodation regulations.</li>
            <li>To communicate with you regarding your booking, rent payments, electricity bills, and other service-related matters.</li>
            <li>To process rent payments, security deposits, and electricity bill settlements.</li>
            <li>To send service notifications, payment reminders, and policy updates.</li>
            <li>To improve our website, services, and user experience based on usage analytics.</li>
            <li>To respond to your inquiries, feedback, or support requests.</li>
            <li>To comply with legal obligations and enforce our Terms &amp; Conditions.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">4. How We Share Your Information</h2>
          <p className="text-on-surface-variant mb-3">
            We do <strong>not</strong> sell, trade, or rent your personal information to third parties. We may share
            information only in the following circumstances:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li><strong>Service Providers:</strong> We may share information with third-party vendors who assist us in operating our website, processing payments, sending communications, or providing analytics. These providers are contractually bound to protect your data.</li>
            <li><strong>Legal Compliance:</strong> We may disclose information if required by law, court order, or government regulation, or to protect the rights, property, or safety of Vats Apartment, our tenants, or others.</li>
            <li><strong>Business Transfers:</strong> In the event of a merger, acquisition, or sale of assets, your information may be transferred as part of the transaction. You will be notified of any such change.</li>
            <li><strong>With Your Consent:</strong> We may share your information for other purposes with your explicit consent.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">5. Data Security</h2>
          <p className="text-on-surface-variant">
            We implement appropriate technical and organizational security measures to protect your
            personal information against unauthorized access, alteration, disclosure, or destruction.
            These measures include:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant mt-3">
            <li>Encrypted data transmission (SSL/TLS) for all website communications.</li>
            <li>Secure database storage with access controls and authentication.</li>
            <li>Session-based authentication for tenant and admin portals.</li>
            <li>Regular security reviews and updates of our systems.</li>
          </ul>
          <p className="text-on-surface-variant mt-3">
            However, no method of transmission over the Internet or electronic storage is 100% secure.
            While we strive to protect your personal information, we cannot guarantee its absolute security.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">6. Data Retention</h2>
          <p className="text-on-surface-variant">
            We retain your personal information only for as long as necessary to fulfill the purposes
            outlined in this Privacy Policy, unless a longer retention period is required or permitted
            by law. Tenant records are retained for the duration of your tenancy plus a reasonable
            period thereafter as required for legal, accounting, and regulatory purposes. When your
            data is no longer needed, we will securely delete or anonymize it.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">7. Cookies and Tracking Technologies</h2>
          <p className="text-on-surface-variant">
            Our website may use cookies and similar tracking technologies to enhance your browsing experience.
            Cookies are small text files stored on your device that help us understand how you use our website.
            You can control cookie preferences through your browser settings. Disabling cookies may affect
            certain features of our website.
          </p>
          <p className="text-on-surface-variant mt-3">
            We may use analytics services (such as Google Analytics) to collect standard internet log
            information and details of visitor behavior patterns. This information is processed in a way
            that does not identify individuals.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">8. Third-Party Links</h2>
          <p className="text-on-surface-variant">
            Our website may contain links to third-party websites or services (such as Google Maps,
            WhatsApp, or payment gateways). We are not responsible for the privacy practices or content
            of these third-party sites. We encourage you to review the privacy policies of any
            third-party services you interact with.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">9. Your Rights</h2>
          <p className="text-on-surface-variant mb-3">
            Under applicable Indian data protection laws (including the Information Technology Act, 2000
            and IT Rules, 2011), you have the following rights regarding your personal information:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li><strong>Access:</strong> You may request a copy of the personal information we hold about you.</li>
            <li><strong>Correction:</strong> You may request correction of inaccurate or incomplete personal data.</li>
            <li><strong>Erasure:</strong> You may request deletion of your personal information, subject to legal retention requirements.</li>
            <li><strong>Restriction:</strong> You may request that we limit the processing of your personal information.</li>
            <li><strong>Objection:</strong> You may object to the processing of your personal information for specific purposes.</li>
            <li><strong>Data Portability:</strong> You may request transfer of your personal data to another service provider.</li>
          </ul>
          <p className="text-on-surface-variant mt-3">
            To exercise any of these rights, please contact us using the details provided in Section 11.
            We will respond to your request within a reasonable timeframe.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">10. Children&apos;s Privacy</h2>
          <p className="text-on-surface-variant">
            Our services are not directed to individuals under the age of 18. We do not knowingly collect
            personal information from children. If we become aware that a child under 18 has provided us
            with personal information, we will take steps to delete such information promptly.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">11. Changes to This Privacy Policy</h2>
          <p className="text-on-surface-variant">
            We reserve the right to update or modify this Privacy Policy at any time. Changes will be
            posted on this page with an updated &quot;Last updated&quot; date. We encourage you to review this
            Privacy Policy periodically. Continued use of our website or services after any changes
            constitutes acceptance of the updated policy.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">12. Contact Us</h2>
          <p className="text-on-surface-variant mb-3">
            If you have any questions, concerns, or requests regarding this Privacy Policy or our data
            practices, please contact us:
          </p>
          <div className="bg-surface-container-low rounded-lg p-5 space-y-2 text-on-surface-variant">
            <p><strong>Vats Apartment</strong></p>
            <p>Near Apna Chai Wala, Lawgate, Jalandhar, Punjab - 144001</p>
            <p>Phone: +91 9992937447</p>
            <p>Email: hament.mailbox@gmail.com</p>
          </div>
        </section>

      </div>
    </main>
  );
}
