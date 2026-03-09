import "../component/Helper.css";
import Sidebar from "./Sidebar";
import loginImg from "../assets/login.png";
import CreatePostImg from "../assets/createpost.png";
import PostDetailImg from "../assets/detail.png";
import AiSumImg from "../assets/ai_sum.png";
import FriendsImg from "../assets/friends.png";
import ProfileImg from "../assets/profile.png";


export default function Helper() {
  return (
    <div className="helper-layout d-flex min-vh-100">
      <Sidebar />

      <main className="helper-main">
        <div className="helper-wrap">

          {/* Header */}
          <header className="helper-header">
            <h2>คู่มือการใช้งาน ChaladShare</h2>
            <p>
              คู่มือแนะนำการใช้งานสำหรับผู้ใช้ อธิบายพร้อมภาพประกอบ 
            </p>
          </header>

          {/* วิธีที่ 1 */}
          <section className="helper-step">
            <figure className="step-figure">
              <h2 className="method-title">1. การเข้าสู่ระบบ หรือสมัครสมาชิก</h2>
              <img src={loginImg} alt="Login" />
              <figcaption>ตัวอย่างหน้าจอเข้าสู่ระบบ และสมัครสมาชิก</figcaption>
            </figure>

            <p className="step-desc">
              ผู้ใช้ต้องเข้าสู่ระบบก่อน เพื่อให้สามารถใช้งานฟีเจอร์ต่างๆของระบบ
              โดยหากเป็นผู้ใช้ใหม่ที่ยังไม่เคยมีบัญชีผู้ใช้ต้องทำการสมัครสมาชิกก่อน
            </p>

            <div className="step-actions">
              <div className="action-item">
                <span className="action-label">Step to Step</span>
                <div className="action-text">
                  ขั้นตอนการ <b>เข้าสู่ระบบ</b> → กรอกอีเมล → กรอกรหัสผ่าน →
                  กดปุ่ม <b>เข้าสู่ระบบ</b>
                </div>

                <div className="action-text">
                  ขั้นตอนการ <b>สมัครสมาชิก</b> → กรอกอีเมล → ตั้งชื่อผู้ใช้ → ตั้งรหัสผ่าน → ยืนยันรหัสผ่าน →
                  กดปุ่ม <b>สมัครสมาชิก</b>
                </div>
              </div>
            </div>
          </section>

          {/* วิธีที่ 2 */}
          <section className="helper-step">
            <figure className="step-figure">
              <h2 className="method-title">2. การดูโพสต์สรุป</h2>
              <img src={PostDetailImg} alt="PostDetail" />
              <figcaption>ตัวอย่างหน้าหลัก</figcaption>
            </figure>

            <p className="step-desc">
              เมื่อผู้ใช้เข้าสู่ระบบจะพบหน้าหลักเป็นหน้าแรก ผู้ใช้สามารถดูโพสต์สรุปที่สนใจจากการคลิกเลือกการ์ด ค้นหาที่ช่องค้นหา และโต้ตอบกับโพสต์สรุปได้
            </p>

            {/* <div className="step-actions">
              <div className="action-item">
                <span className="action-label">Step to Step</span>
                <div className="action-text">
                  ขั้นตอนการ <b>เข้าสู่ระบบ</b> → กรอกอีเมล → กรอกรหัสผ่าน →
                  กดปุ่ม <b>เข้าสู่ระบบ</b>
                </div>

                <div className="action-text">
                  ขั้นตอนการ <b>สมัครสมาชิก</b> → กรอกอีเมล → ตั้งชื่อผู้ใช้ → ตั้งรหัสผ่าน → ยืนยันรหัสผ่าน →
                  กดปุ่ม <b>สมัครสมาชิก</b>
                </div>
              </div>
            </div> */}
          </section>

          {/* วิธีที่ 3 */}
          <section className="helper-step">
            <figure className="step-figure">
              <h2 className="method-title">3. การสร้างและอัพโหลดโพสต์สรุปใหม่เข้าสู่ระบบ</h2>
              <img src={CreatePostImg} alt="Create Post" />
              <figcaption>ตัวอย่างหน้าจอการสร้างโพสต์สรุป</figcaption>
            </figure>

            <p className="step-desc">
              ผู้ใช้สามารถอัปโหลดโพสต์สรุปใหม่เข้าสู่ระบบ
              โดยคลิกเลือกที่ปุ่ม "สร้าง" ที่เถบเมนูด้านซ้าย
            </p>

            <div className="step-actions">
              <div className="action-item">
                <span className="action-label">Step to Step</span>
                <div className="action-text">
                  ขั้นตอนการ <b>โพสต์สรุป</b> → กรอกหัวข้อ → ตั้งค่าระดับการมองเห็น → อัปโหลดรูปภาพหน้าปก → อัปโหลดไฟล์สรุป → ใส่คำอธิบายโพสต์ (ถ้ามี) → ใส่แฮชแท็ก →
                  กดปุ่ม <b>โพสต์</b>
                </div>

                {/* <div className="action-text">
                  ขั้นตอนการ <b>สมัครสมาชิก</b> → กรอกอีเมล → ตั้งชื่อผู้ใช้ → ตั้งรหัสผ่าน → ยืนยันรหัสผ่าน →
                  กดปุ่ม <b>สมัครสมาชิก</b>
                </div> */}
              </div>
            </div>
          </section>

          {/* วิธีที่ 4 */}
          <section className="helper-step">
            <figure className="step-figure">
              <h2 className="method-title">4. การใช้งาน AI ช่วยสรุป</h2>
              <img src={AiSumImg} alt="AI Summaryt" />
              <figcaption>ตัวอย่างหน้าจอ AI ช่วยสรุป</figcaption>
            </figure>

            <p className="step-desc">
              ผู้ใช้สามารถใช้งานการสรุปเนื้อหาอัตโนมัติด้วย AI
              โดยคลิกเลือกที่ปุ่ม "AI ช่วยสรุป" ที่เถบเมนูด้านซ้าย
            </p>

            <div className="step-actions">
              <div className="action-item">
                <span className="action-label">Step to Step</span>
                <div className="action-text">
                  ขั้นตอนการ <b>สรุปเนื้อหาอัตโนมัติ</b> → กดปุ่มเพิ่มไฟล์ → เลือกไฟล์ที่ต้องการ → รอผลลัพธ์สักครู่
                </div>

                {/* <div className="action-text">
                  ขั้นตอนการ <b>สมัครสมาชิก</b> → กรอกอีเมล → ตั้งชื่อผู้ใช้ → ตั้งรหัสผ่าน → ยืนยันรหัสผ่าน →
                  กดปุ่ม <b>สมัครสมาชิก</b>
                </div> */}
              </div>
            </div>
          </section>

          {/* วิธีที่ 5*/}
          <section className="helper-step">
            <figure className="step-figure">
              <h2 className="method-title">5. การจัดการรายการเพื่อน</h2>
              <img src={FriendsImg} alt="Friends" />
              <figcaption>ตัวอย่างหน้าจอรายการเพื่อน</figcaption>
            </figure>

            <p className="step-desc">
              ผู้ใช้สามารถดูและลบรายการเพื่อนของตนเอง รวมถึงการจัดการคำขอเป็นเพื่อน (ตอบรับ/ปฏิเสธ) 
             
            </p>

            {/* <div className="step-actions">
              <div className="action-item">
                <span className="action-label">Step to Step</span>
                <div className="action-text">
                  ขั้นตอนการ <b>สรุปเนื้อหาอัตโนมัติ</b> → กดปุ่มเพิ่มไฟล์ → เลือกไฟล์ที่ต้องการ → รอผลลัพธ์สักครู่
                </div>
              </div>
            </div> */}
          </section>


          {/* วิธีที่ 6*/}
          <section className="helper-step">
            <figure className="step-figure">
              <h2 className="method-title">6. การจัดการโปรไฟล์</h2>
              <img src={ProfileImg} alt="Profile" />
              <figcaption>ตัวอย่างหน้าจอโปรไฟล์ของฉัน</figcaption>
            </figure>

            <p className="step-desc">
              ในหน้านี้จะแสดงข้อมูล และโพสต์ที่ผู้ใช้ได้ทำการอัปโหลด
              โดยผู้ใช้สามารถจัดการข้อมูลของตนเองที่ปุ่ม "แก้ไขโปรไฟล์" 
              ดูรายการโพสต์ที่ตนบันทึกไว้ที่ปุ่ม "รายการที่บันทึก" 
              และจัดการโพสต์ของตนเอง ( แก้ไข / ลบ ) ด้านล่างมุมขวาของการ์ด
              
            </p>
{/* 
            <div className="step-actions">
              <div className="action-item">
                <span className="action-label">Step to Step</span>
                <div className="action-text">
                  ขั้นตอนการ <b>สรุปเนื้อหาอัตโนมัติ</b> → กดปุ่มเพิ่มไฟล์ → เลือกไฟล์ที่ต้องการ → รอผลลัพธ์สักครู่
                </div>
              </div>
            </div> */}
          </section>


        </div>
      </main>
    </div>
  );
}