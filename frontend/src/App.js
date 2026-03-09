import "./App.css";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";

import Login from "./pages/Login";
import Register from "./pages/Register";
import Home from "./pages/Home";
import PostDetail from "./pages/PostDetail";
import CreatePost from "./pages/Createpost";
import Friends from "./pages/Friends";
import Profile from "./pages/Profile";
import ForgotPassword from "./pages/ForgotPassword";
import EditPost from "./pages/EditPost";
import AISummary from "./pages/AISummary";
import Helper from "./pages/Helper";
import VerifyOTP from "./pages/VerifyOTP";
import NewPassword from "./pages/NewPassword";
import RequireAuth from "./component/RequireAuth";

function App() {
  return (
    <Router>
      <Routes>
        {/* <Route path="/" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/home" element={<Home />} />
        <Route path="/posts/:id" element={<PostDetail />} />
        <Route path="/newpost" element={<CreatePost />} />
        <Route path="/friends" element={<Friends />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/profile/:id" element={<Profile />} />
        <Route path="/forgot_password" element={<ForgotPassword />} />
        <Route path="/posts/:id/edit" element={<EditPost />} />
        <Route path="/ai" element={<AISummary />} />
        <Route path="/helper" element={<Helper />} />
        <Route path="/verify-otp" element={<VerifyOTP />} />
        <Route path="/new-password" element={<NewPassword />} /> */}
        <Route path="/" element={<Login />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/forgot_password" element={<ForgotPassword />} />
        <Route path="/verify-otp" element={<VerifyOTP />} />
        <Route path="/new-password" element={<NewPassword />} />

        <Route element={<RequireAuth />}>
          <Route path="/home" element={<Home />} />
          <Route path="/newpost" element={<CreatePost />} />
          <Route path="/ai" element={<AISummary />} />
          <Route path="/friends" element={<Friends />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/profile/:id" element={<Profile />} />
          <Route path="/posts/:id" element={<PostDetail />} />
          <Route path="/posts/:id/edit" element={<EditPost />} />
          <Route path="/helper" element={<Helper />} />
        </Route>
      </Routes>
    </Router>
  );
}

export default App;
