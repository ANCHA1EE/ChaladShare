import React from 'react';

export default function Footer() {
    const SIDEBAR_WIDTH = 260;
    return (
        <footer style={{ backgroundColor: "#026ba5", color: "white", display: "flex", justifyContent: "center", width:"100%", padding: "15px 0" }}>
            <div style={{ marginLeft: SIDEBAR_WIDTH, width: `calc(100% - ${SIDEBAR_WIDTH}px)`, textAlign: "center" }}>
                 <p style={{ margin: "0", color: "white" }}>&copy; 2025 Chalad Share</p>
                 <p style={{ margin: "0", color: "white" }}>เว็บไซต์นี้จัดทำขึ้นเพื่อโครงงานปริญญานิพนธ์ </p>
            </div>
        </footer>
    );
}

