import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Container,
  Paper,
  TextField,
  Button,
  Typography,
  Alert,
  IconButton,
  InputAdornment,
} from "@mui/material";
import { EyeOutlined, EyeInvisibleOutlined } from "@ant-design/icons";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const formData = new FormData();
      formData.append("username", username);
      formData.append("password", password);

      const response = await fetch("/api/v1/admin/login", {
        method: "POST",
        body: formData,
      });

      if (response.ok) {
        navigate("/");
      } else {
        const data = await response.json().catch(() => ({}));
        setError(data.error || "Invalid credentials");
      }
    } catch (err) {
      setError("Unable to connect to server");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        bgcolor: "background.default",
      }}
    >
      <Container maxWidth="sm">
        <Paper
          sx={{ p: { md: 6, xs: 3 }, maxWidth: 500, mx: "auto" }}
          variant="outlined"
        >
          <Typography variant="h5" align="center" gutterBottom fontWeight={600}>
            ShortMesh Admin
          </Typography>
          <Typography
            variant="body2"
            align="center"
            color="text.secondary"
            sx={{ mb: 4 }}
          >
            Sign in to manage your system
          </Typography>
          <form onSubmit={handleSubmit}>
            <TextField
              fullWidth
              variant="filled"
              label="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
              required
              autoFocus
            />
            <TextField
              fullWidth
              variant="filled"
              label="Password"
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
              required
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={() => setShowPassword(!showPassword)}
                      onMouseDown={(e) => e.preventDefault()}
                      edge="end"
                      sx={{ mr: 0.01 }}
                    >
                      {showPassword ? (
                        <EyeInvisibleOutlined style={{ fontSize: 18 }} />
                      ) : (
                        <EyeOutlined style={{ fontSize: 18 }} />
                      )}
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
            <Button
              fullWidth
              type="submit"
              variant="contained"
              size="large"
              disabled={loading}
              sx={{ mt: 3 }}
            >
              {loading ? "Signing In..." : "Sign In"}
            </Button>
          </form>
          {error && (
            <Alert severity="error" sx={{ mt: 2, py: 2 }}>
              {error}
            </Alert>
          )}
        </Paper>
      </Container>
    </Box>
  );
}
