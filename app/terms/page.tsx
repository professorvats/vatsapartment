import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms & Conditions",
  description: "Terms and Conditions for Vats Apartment - booking, rent, security deposit, electricity billing, and tenant obligations.",
  alternates: { canonical: 'https://vatsapartment.com/terms' },
  robots: { index: true, follow: true },
};

export default function TermsPage() {
  return (
    <main className="container-max gutter py-12 md:py-16 font-sans max-w-4xl mx-auto">
      <h1 className="text-3xl md:text-4xl font-bold text-on-surface mb-2">
        Terms &amp; Conditions
      </h1>
      <p className="text-on-surface-variant text-sm mb-10">
        Last updated: June 2026
      </p>

      <div className="prose prose-neutral max-w-none space-y-8 text-on-surface text-sm md:text-base leading-relaxed">

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">1. Acceptance of Terms</h2>
          <p className="text-on-surface-variant">
            By accessing or using the Vats Apartment website (<strong>vatsapartment.com</strong>), booking a room,
            or entering into a tenancy agreement with us, you agree to be bound by these Terms &amp; Conditions.
            If you do not agree with any part of these terms, you must not use our services or website.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">2. Definitions</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li><strong>&quot;We&quot;, &quot;Us&quot;, &quot;Our&quot;:</strong> Vats Apartment, the property management and accommodation provider.</li>
            <li><strong>&quot;You&quot;, &quot;Tenant&quot;, &quot;Occupant&quot;:</strong> The individual booking or occupying a room at Vats Apartment.</li>
            <li><strong>&quot;Property&quot;:</strong> The residential premises located at Near Apna Chai Wala, Lawgate, Jalandhar, Punjab.</li>
            <li><strong>&quot;Room&quot;:</strong> A specific 1 BHK unit within the Property assigned to the tenant.</li>
            <li><strong>&quot;Security Deposit&quot;:</strong> A refundable amount collected at the start of tenancy, equivalent to 10 (ten) months&apos; rent, adjusted solely against outstanding electricity bills at the end of the tenancy.</li>
            <li><strong>&quot;Booking&quot;:</strong> A confirmed reservation for a room at the Property.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">3. Booking and Reservation</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>All bookings are subject to room availability and management approval.</li>
            <li>A booking is confirmed only upon payment of the security deposit and the first month&apos;s rent.</li>
            <li>Booking confirmations are communicated via phone, WhatsApp, or email.</li>
            <li>We reserve the right to cancel any booking at our discretion, with a full refund of amounts paid.</li>
            <li>You must provide accurate and complete personal information, including government-issued ID proof, during the booking process.</li>
            <li>The minimum rental commitment is month-to-month unless otherwise agreed in writing.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">4. Rent and Payment Terms</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>Monthly rent is due on or before the <strong>1st day of each month</strong>.</li>
            <li>Rent payments shall be made via accepted payment methods as communicated by management (bank transfer, UPI, or cash).</li>
            <li>A <strong>grace period of 5 days</strong> is provided after the due date.</li>
            <li>Late payment beyond the grace period incurs a <strong>late fee of 10% per month</strong> on the outstanding rent amount.</li>
            <li>Rent amounts may be revised by management with <strong>30 days&apos; written notice</strong> to the tenant.</li>
            <li>Partial payments are accepted only at management&apos;s discretion and do not waive the full rent obligation.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">5. Security Deposit</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>A refundable security deposit of <strong>10 (ten) months&apos; rent</strong> is required at the start of the tenancy.</li>
            <li><strong>There is NO lock-in period.</strong> The security deposit is fully refundable at the end of tenancy.</li>
            <li>The security deposit <strong>shall be adjusted solely against outstanding electricity bills</strong> at the time of vacating the premises.</li>
            <li>No deductions shall be made from the security deposit for any other purpose (including rent arrears, damages, or other charges) except electricity bills.</li>
            <li>Upon vacating the property and settling all electricity bills, the remaining security deposit balance shall be refunded within <strong>15 working days</strong>.</li>
            <li>Any dispute regarding security deposit deductions shall be resolved through mutual discussion and reconciliation of electricity bill records.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">6. Electricity Billing</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>Electricity charges are <strong>NOT included</strong> in the monthly rent and shall be paid separately by the tenant.</li>
            <li>Each room is equipped with a sub-meter for individual electricity consumption tracking.</li>
            <li>Meter readings are taken periodically by management and shared with the tenant for verification.</li>
            <li>The electricity rate is charged at <strong>₹12 per unit</strong> (subject to revision based on actual utility rates).</li>
            <li>Electricity bills are generated monthly and are payable within <strong>7 days</strong> of issuance.</li>
            <li>Unpaid electricity bills at the end of tenancy shall be deducted from the security deposit as per Section 5.</li>
            <li>Tenants have the right to review and verify meter readings and bill calculations through their tenant portal.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">7. Tenant Obligations</h2>
          <p className="text-on-surface-variant mb-3">As a tenant, you agree to:</p>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>Use the room solely for residential purposes. Commercial or business activities are strictly prohibited.</li>
            <li>Maintain the room in clean and habitable condition at all times.</li>
            <li>Not cause any damage to the property, fixtures, furniture, appliances, or common areas.</li>
            <li>Not sublet, assign, or transfer the room to any other person without prior written consent from management.</li>
            <li>Not make any structural alterations, installations, or modifications to the room without prior written approval.</li>
            <li>Comply with all applicable laws, building rules, and community regulations.</li>
            <li>Report any maintenance issues, damages, or safety concerns to management promptly.</li>
            <li>Use electricity, water, and other utilities responsibly and avoid wastage.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">8. Property Rules</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li><strong>Quiet Hours:</strong> 10:00 PM to 7:00 AM. Excessive noise, loud music, or disturbances during quiet hours are prohibited.</li>
            <li><strong>Visitors:</strong> Visitors are permitted during reasonable hours only. Overnight guests require prior notification to management. The tenant is responsible for their visitors&apos; conduct.</li>
            <li><strong>Smoking &amp; Alcohol:</strong> Smoking and consumption of alcohol are strictly prohibited inside the rooms and common areas of the property.</li>
            <li><strong>Pets:</strong> Pets are not permitted unless explicitly approved in writing by management.</li>
            <li><strong>Parking:</strong> Designated parking spaces, if available, must be used. Unauthorized parking may result in vehicle removal at the owner&apos;s expense.</li>
            <li><strong>Common Areas:</strong> Common areas (corridors, stairs, terrace, etc.) must be kept clean and free of personal belongings.</li>
            <li><strong>Security:</strong> The property is under 24/7 CCTV surveillance in common areas. Tampering with security equipment is a serious violation.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">9. Termination and Vacating</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>Either party may terminate the tenancy by providing <strong>30 days&apos; written notice</strong>.</li>
            <li>Upon termination, the tenant must vacate the room and return all keys, access cards, and property items.</li>
            <li>The room must be returned in the same condition as at the start of the tenancy, subject to normal wear and tear.</li>
            <li>A final inspection will be conducted jointly by management and the tenant before vacating.</li>
            <li>Management reserves the right to terminate tenancy immediately in cases of:
              <ul className="list-[circle] pl-6 mt-1 space-y-1">
                <li>Non-payment of rent for more than 30 days beyond the grace period.</li>
                <li>Illegal activities or violation of law on the premises.</li>
                <li>Repeated or serious violation of property rules.</li>
                <li>Causing significant damage to the property or endangering other occupants.</li>
              </ul>
            </li>
            <li>In case of immediate termination for cause, the security deposit shall still be refunded as per Section 5 after adjustment of electricity bills.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">10. Damages and Liability</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>The tenant is responsible for any damage caused to the room, furniture, appliances, fixtures, or common areas beyond normal wear and tear.</li>
            <li>Damage repair costs will be communicated to the tenant and must be settled promptly. Damages are <strong>separate from the security deposit</strong> and must be paid additionally.</li>
            <li>Vats Apartment is not liable for loss, theft, or damage to the tenant&apos;s personal belongings. Tenants are advised to secure their valuables.</li>
            <li>Vats Apartment is not liable for any injury, accident, or incident occurring on the property unless caused by our proven negligence.</li>
            <li>Tenants are encouraged to obtain personal insurance for their belongings if desired.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">11. Maintenance and Repairs</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>Management is responsible for structural repairs and maintenance of major appliances and fixtures provided with the room.</li>
            <li>Tenants must promptly report any maintenance issues through the tenant portal, phone, or WhatsApp.</li>
            <li>Management will address reported issues within a reasonable timeframe based on severity.</li>
            <li>Minor maintenance (e.g., changing light bulbs, cleaning) is the tenant&apos;s responsibility.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">12. Intellectual Property</h2>
          <p className="text-on-surface-variant">
            All content on the Vats Apartment website, including text, graphics, logos, images, and software,
            is the property of Vats Apartment and is protected by applicable intellectual property laws.
            You may not reproduce, distribute, or use any content from our website without prior written permission.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">13. Disclaimer of Warranties</h2>
          <p className="text-on-surface-variant">
            Our website and services are provided on an &quot;as is&quot; and &quot;as available&quot; basis. We make no
            warranties, expressed or implied, regarding the website&apos;s operation, availability, or accuracy
            of information. We do not guarantee uninterrupted or error-free service.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">14. Limitation of Liability</h2>
          <p className="text-on-surface-variant">
            To the maximum extent permitted by applicable law, Vats Apartment shall not be liable for any
            indirect, incidental, special, consequential, or punitive damages arising out of or relating
            to your use of our website, services, or tenancy, whether based on warranty, contract, tort,
            or any other legal theory.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">15. Governing Law and Dispute Resolution</h2>
          <ul className="list-disc pl-6 space-y-2 text-on-surface-variant">
            <li>These Terms &amp; Conditions shall be governed by and construed in accordance with the laws of India.</li>
            <li>Any dispute arising out of or relating to these terms shall first be attempted to be resolved through mutual negotiation and discussion between the parties.</li>
            <li>If a resolution cannot be reached through negotiation, the dispute shall be subject to the exclusive jurisdiction of the courts in Jalandhar, Punjab.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">16. Changes to Terms</h2>
          <p className="text-on-surface-variant">
            We reserve the right to modify or update these Terms &amp; Conditions at any time. Changes will
            be posted on this page with an updated &quot;Last updated&quot; date. Continued use of our services
            after any changes constitutes acceptance of the updated terms. For significant changes affecting
            existing tenancies, we will provide direct notice to affected tenants.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-on-surface mb-3">17. Contact Information</h2>
          <p className="text-on-surface-variant mb-3">
            For any questions, concerns, or clarifications regarding these Terms &amp; Conditions, please contact:
          </p>
          <div className="bg-surface-container-low rounded-lg p-5 space-y-2 text-on-surface-variant">
            <p><strong>Vats Apartment</strong></p>
            <p>Near Apna Chai Wala, Lawgate, Jalandhar, Punjab - 144001</p>
            <p>Phone: +91 9992937447</p>
            <p>Email: hament.mailbox@gmail.com</p>
            <p>WhatsApp: +91 9992937447</p>
          </div>
        </section>

      </div>
    </main>
  );
}
