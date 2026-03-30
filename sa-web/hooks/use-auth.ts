"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { authService } from "@/services/auth.service";
import { LoginRequest } from "@/types/auth";
import { getStoredUser } from "@/lib/auth";

export function useCurrentUser() {
  return useQuery({
    queryKey: ["current-user"],
    queryFn: () => authService.getMe(),
    initialData: getStoredUser() ?? undefined,
    staleTime: 5 * 60 * 1000,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (result) => {
      queryClient.setQueryData(["current-user"], result.user);
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return () => {
    queryClient.clear();
    authService.logout();
  };
}
